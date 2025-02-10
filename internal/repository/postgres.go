package repository

import (
	"context"
	"fmt"
	"reflect"

	"github.com/22Fariz22/forum/internal/model"
	"github.com/jmoiron/sqlx"
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
		return fmt.Errorf("failed to create user: %w", err) // 500 Internal Server Error
	}

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
