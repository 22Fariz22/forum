package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"reflect"

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

func (r *PostgresRepository) CreateUser(user *model.User) error {
	fmt.Println("in repo pg CreateUser(). username:", user.Username)
	fmt.Println("id type of:", reflect.TypeOf(user.ID))
	// Проверяем, что username не пустой (это уже должно быть проверено в резолвере)
	if user.Username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	// SQL-запрос для вставки пользователя
	query := `
		INSERT INTO users (id, username)
		VALUES ($1, $2)
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

func (r *PostgresRepository) CreatePost(post *model.Post) error {
	fmt.Println("in repo pg CreatePost()")
	fmt.Println("post:", post)
	// SQL-запрос для вставки поста
	query := `
		INSERT INTO posts (id, title, content, allow_comments, have_comments, author_id)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	// Выполняем запрос
	_, err := r.db.Exec(query, post.ID, post.Title, post.Content, post.AllowComments, post.HaveComments, post.AuthorID)
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

func (r *PostgresRepository) GetPosts() ([]*model.Post, error) {
	fmt.Println("in repo pg GetPosts()")
	query := `
		SELECT id, title, content, allow_comments, author_id, have_comments
		FROM posts
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

func (r *PostgresRepository) GetPostByID(id string) (*model.Post, error) {
	fmt.Println("in repo pg GetPostByID()")

	query := `
		SELECT id, title, content, allow_comments, author_id, have_comments
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

func (r *PostgresRepository) CreateCommentOnPost(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	fmt.Printf("in repo CreateCommentOnPost")
	fmt.Println("comment from resolve: ", comment)
	fmt.Println("comment.Author:", comment.Author)

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

func (r *PostgresRepository) ReplyToComment(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	return nil, nil
}

func (r *PostgresRepository) GetCommentsByPostID(postID string, limit, offset int) ([]*model.Comment, error) {
	fmt.Println("in repo GetCommentsByPostID. postID:", postID, " limit:", limit, " offset:", offset)

	// SQL-запрос для получения комментариев с пагинацией
	query := `
		SELECT id, post_id, parent_id, content, author_id, username, have_comments
		FROM comments
		WHERE post_id = $1
		ORDER BY id DESC
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

func (r *PostgresRepository) GetReplies(parentID string) ([]*model.Comment, error) {
	return nil, nil
}

func isDuplicateKeyError(err error) bool {
	// PostgreSQL возвращает ошибку с кодом "23505" при нарушении уникальности
	if err, ok := err.(*pq.Error); ok {
		return err.Code == "23505"
	}
	return false
}
