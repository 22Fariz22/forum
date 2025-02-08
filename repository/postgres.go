package repository

import (
	"database/sql"
	"errors"

	"github.com/22Fariz22/forum/internal/model"
	// "github.com/lib/pq"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(connStr string) (Repository, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &PostgresRepository{db: db}, nil
}

func (r *PostgresRepository) CreatePost(post *model.Post) error {
	_, err := r.db.Exec("INSERT INTO posts (id, title, content, allow_comments) VALUES ($1, $2, $3, $4)",
		post.ID, post.Title, post.Content, post.AllowComments)
	return err
}

func (r *PostgresRepository) GetPosts() ([]*model.Post, error) {
	rows, err := r.db.Query("SELECT id, title, content, allow_comments FROM posts")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []*model.Post
	for rows.Next() {
		var post model.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.AllowComments); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}
	return posts, nil
}

func (r *PostgresRepository) GetPostByID(id string) (*model.Post, error) {
	row := r.db.QueryRow("SELECT id, title, content, allow_comments FROM posts WHERE id = $1", id)
	var post model.Post
	if err := row.Scan(&post.ID, &post.Title, &post.Content, &post.AllowComments); err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("пост не найден")
		}
		return nil, err
	}
	return &post, nil
}

func (r *PostgresRepository) CreateComment(comment *model.Comment) error {
	_, err := r.db.Exec("INSERT INTO comments (id, post_id, parent_id, content, author) VALUES ($1, $2, $3, $4, $5)",
		comment.ID, comment.PostID, comment.ParentID, comment.Content, comment.Author)
	return err
}

func (r *PostgresRepository) GetCommentsByPostID(postID string, limit, offset int) ([]*model.Comment, error) {
	rows, err := r.db.Query("SELECT id, post_id, parent_id, content, author FROM comments WHERE post_id = $1 AND parent_id IS NULL ORDER BY id LIMIT $2 OFFSET $3", postID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []*model.Comment
	for rows.Next() {
		var comment model.Comment
		var parentID sql.NullString
		if err := rows.Scan(&comment.ID, &comment.PostID, &parentID, &comment.Content, &comment.Author); err != nil {
			return nil, err
		}
		if parentID.Valid {
			comment.ParentID = &parentID.String
		}
		comments = append(comments, &comment)
	}
	return comments, nil
}

func (r *PostgresRepository) GetCommentsByParentID(parentID string, limit, offset int) ([]*model.Comment, error) {
	rows, err := r.db.Query("SELECT id, post_id, parent_id, content, author FROM comments WHERE parent_id = $1 ORDER BY id LIMIT $2 OFFSET $3", parentID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var comments []*model.Comment
	for rows.Next() {
		var comment model.Comment
		var pid sql.NullString
		if err := rows.Scan(&comment.ID, &comment.PostID, &pid, &comment.Content, &comment.Author); err != nil {
			return nil, err
		}
		if pid.Valid {
			comment.ParentID = &pid.String
		}
		comments = append(comments, &comment)
	}
	return comments, nil
}
