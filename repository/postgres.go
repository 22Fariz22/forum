package repository

import (
	"context"
	"database/sql"

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

func (r *PostgresRepository) CreateUser(user *model.User) error {
	return nil
}

func (r *PostgresRepository) GetUserByID(id string) (*model.User, error) {
	return nil, nil
}

func (r *PostgresRepository) CreatePost(post *model.Post) error {
	return nil
}

func (r *PostgresRepository) GetPosts() ([]*model.Post, error) {
	return nil, nil
}

func (r *PostgresRepository) GetPostByID(id string) (*model.Post, error) {

	return nil, nil
}

func (r *PostgresRepository) CreateCommentOnPost(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	return nil, nil
}
func (r *PostgresRepository) ReplyToComment(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	return nil, nil
}

func (r *PostgresRepository) GetCommentsByPostID(postID string, limit, offset int) ([]*model.Comment, error) {
	return nil, nil
}

func (r *PostgresRepository) GetReplies(parentID string) ([]*model.Comment, error) {
	return nil, nil
}

func (r *PostgresRepository) SeedData() {}
