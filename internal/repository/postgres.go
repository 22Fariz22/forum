package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	// "time"

	"github.com/22Fariz22/forum/internal/model"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type PostgresRepository struct {
	db *sqlx.DB
}

func NewPostgresRepository(db *sqlx.DB) (Repository, error) {
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresRepository{db: db}, nil
}

// CreateUser создаем пользователя
func (r *PostgresRepository) CreateUser(user *model.User) error {
	// Проверяем, что username не пустой (это уже должно быть проверено в резолвере)
	if user.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// SQL-запрос для вставки пользователя
	query := `
		INSERT INTO users (id, username, created_at)
		VALUES ($1, $2, NOW())
	`

	// Выполняем запрос
	_, err := r.db.Exec(query, user.ID, user.Username)
	if err != nil {
		// Обрабатываем ошибки
		switch {
		case isDuplicateKeyError(err):
			return fmt.Errorf("username already exists") // 409 Conflict
		default:
			return fmt.Errorf("failed to create user: %w", err) // 500 Internal Server Error
		}
	}

	return nil
}

// GetUserByID получаем пользователя по ID
func (r *PostgresRepository) GetUserByID(id string) (*model.User, error) {
	query := `SELECT id, username FROM users WHERE id = $1`

	var user model.User
	err := r.db.QueryRowContext(context.Background(), query, id).Scan(&user.ID, &user.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("пользователь не найден")
		}
		return nil, err
	}

	return &user, nil
}

// CreatePost создаем пост
func (r *PostgresRepository) CreatePost(post *model.Post) error {
	// SQL-запрос для вставки поста
	query := `
		INSERT INTO posts (id, title, content, allow_comments, have_comments, author_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	// Выполняем запрос
	_, err := r.db.Exec(
		query,
		post.ID,
		post.Title,
		post.Content,
		post.AllowComments,
		post.HaveComments,
		post.AuthorID,
	)
	if err != nil {
		// Обрабатываем ошибки
		switch {
		case isDuplicateKeyError(err):
			return fmt.Errorf("post with this ID already exists") // 409 Conflict
		default:
			return fmt.Errorf("failed to create post: %w", err) // 500 Internal Server Error
		}
	}

	return nil
}

// GetPosts получаем все посты
func (r *PostgresRepository) GetPosts(offset int32, limit int32) ([]*model.Post, error) {
	query := `
		SELECT id, title, content, allow_comments, author_id, have_comments, created_at
		FROM posts
		ORDER BY created_at ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts: %w", err)
	}
	defer rows.Close()

	var posts []*model.Post
	for rows.Next() {
		post := &model.Post{}
		err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Content,
			&post.AllowComments,
			&post.AuthorID,
			&post.HaveComments,
			&post.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan post: %w", err)
		}
		posts = append(posts, post)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over posts: %w", err)
	}

	return posts, nil
}

// GetPostByID получаем конкретный пост по post_id
func (r *PostgresRepository) GetPostByID(id string) (*model.Post, error) {
	query := `
		SELECT id, title, content, allow_comments, author_id, have_comments, created_at
		FROM posts
		WHERE id = $1
	`

	post := &model.Post{}
	err := r.db.QueryRow(query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.AllowComments,
		&post.AuthorID,
		&post.HaveComments,
		&post.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("post not found: %w", err) // Ошибка 404
		}
		return nil, fmt.Errorf("failed to fetch post: %w", err)
	}

	// Теперь получаем комментарии для поста
	comments, err := r.GetCommentsByPostID(post.ID, 10, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %w", err)
	}

	post.Comments = comments

	return post, nil
}

// CreateCommentOnPost создаем верхнеуровневый коментарий к посту
func (r *PostgresRepository) CreateCommentOnPost(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	query := `
    INSERT INTO comments (id, post_id, parent_id, content, author_id,username, have_comments) 
    VALUES ($1, $2, $3, $4, $5, $6, $7)
`
	_, err := r.db.ExecContext(
		ctx,
		query,
		comment.ID,
		comment.PostID,
		comment.ParentID,
		comment.Content,
		comment.Author.ID,
		comment.Author.Username,
		comment.HaveComments)
	if err != nil {
		return nil, err
	}

	return comment, nil
}

// ReplyToComment создаем вложенный коментарий, то есть ответ на коментарий
func (r *PostgresRepository) ReplyToComment(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	// Проверяем, существует ли родительский комментарий
	var parentComment model.Comment
	query := `
		SELECT id, post_id AS "post_id"
		FROM comments
		WHERE id = $1
	`

	err := r.db.GetContext(ctx, &parentComment, query, *comment.ParentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			fmt.Println("err in db.GetContext. error.Is", err)
			return nil, fmt.Errorf("родительский комментарий не найден") // 404 Not Found
		}
		fmt.Println("err in db.GetContext", err)
		return nil, fmt.Errorf("failed to fetch parent comment: %w", err) // 500 Internal Server Error
	}

	// Устанавливаем post_id для нового комментария
	comment.PostID = parentComment.PostID

	// SQL-запрос для создания нового комментария
	insertQuery := `
		INSERT INTO comments (id, post_id, parent_id, content, author_id, username)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	// Выполняем запрос на вставку
	_, err = r.db.ExecContext(ctx, insertQuery,
		comment.ID,
		comment.PostID,
		comment.ParentID,
		// parentID,
		comment.Content,
		comment.Author.ID,
		comment.Author.Username,
	)
	if err != nil {
		fmt.Println("err in  r.db.ExecContext:", err)
		return nil, fmt.Errorf("failed to create comment: %w", err) // 500 Internal Server Error
	}

	// Обновляем флаг have_comments у родительского комментария
	updateQuery := `
		UPDATE comments
		SET have_comments = true
		WHERE id = $1
	`
	_, err = r.db.ExecContext(ctx, updateQuery, *comment.ParentID)
	if err != nil {
		fmt.Println("r.db.ExecContext(ctx, updateQuery, *comment.ParentID) err:", err)
		return nil, fmt.Errorf("failed to update parent comment: %w", err) // 500 Internal Server Error
	}

	// Возвращаем созданный комментарий
	return comment, nil
}

// GetCommentsByPostID получаем верхнеуровневые коментарии к посту используя пагинацию
func (r *PostgresRepository) GetCommentsByPostID(postID string, limit, offset int) ([]*model.Comment, error) {
	// SQL-запрос для получения комментариев с пагинацией
	query := `
		SELECT id, post_id, parent_id, content, author_id, username, have_comments, created_at
		FROM comments
		WHERE post_id = $1 and parent_id IS NULL 
		ORDER BY created_at ASC
		LIMIT $2 OFFSET $3
	`

	// Выполняем запрос
	rows, err := r.db.Query(query, postID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %w", err)
	}
	defer rows.Close()

	// Сканируем результаты в структуру Comment
	var comments []*model.Comment
	for rows.Next() {
		var comment model.Comment
		var parentID *string

		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&parentID,
			&comment.Content,
			&comment.AuthorID,
			&comment.Username,
			&comment.HaveComments,
			&comment.CreatedAt,
		)
		if err != nil {
			fmt.Printf("Error during scan: %v\n", err)
			return nil, fmt.Errorf("failed to scan comment: %w", err)
		}

		author := &model.User{
			ID:       comment.AuthorID,
			Username: comment.Username,
		}
		comment.Author = author

		comments = append(comments, &comment)
	}

	// Проверяем, не возникла ли ошибка при итерации по строкам
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return comments, nil
}

// GetReplies получение вложенных комментариев по id родительского коментария
func (r *PostgresRepository) GetReplies(parentID string) ([]*model.Comment, error) {
	// SQL-запрос для получения вложенных комментариев
	query := `
		SELECT id, post_id, parent_id, content, author_id, username, have_comments, created_at
		FROM comments
		WHERE parent_id = $1 
		ORDER BY created_at ASC
	`

	// Выполняем запрос
	rows, err := r.db.Query(query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch replies: %w", err)
	}
	defer rows.Close()

	// Сканируем результаты в структуру Comment
	var replies []*model.Comment
	for rows.Next() {
		var comment model.Comment

		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.ParentID,
			&comment.Content,
			&comment.AuthorID,
			&comment.Username,
			&comment.HaveComments,
			&comment.CreatedAt,
		)
		if err != nil {
			fmt.Printf("Error during scan: %v\n", err)
			return nil, fmt.Errorf("failed to scan reply: %w", err)
		}

		// Создаем объект автора
		author := &model.User{
			ID:       comment.AuthorID,
			Username: comment.Username,
		}
		comment.Author = author

		replies = append(replies, &comment)
	}

	// Проверяем, не возникла ли ошибка при итерации по строкам
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return replies, nil
}

// isDuplicateKeyError проверка дупликата
func isDuplicateKeyError(err error) bool {
	// PostgreSQL возвращает ошибку с кодом "23505" при нарушении уникальности
	if err, ok := err.(*pq.Error); ok {
		return err.Code == "23505"
	}
	return false
}
