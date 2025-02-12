package repository

import (
	"context"

	"github.com/22Fariz22/forum/internal/model"
)

// Repository – интерфейс для работы с постами и комментариями
type Repository interface {
	//методы для пользователя
	CreateUser(user *model.User) error
	GetUserByID(id string) (*model.User, error)

	// Методы для постов
	CreatePost(post *model.Post) error
	GetPosts(offset int32, limit int32) ([]*model.Post, error)
	GetPostByID(id string) (*model.Post, error)

	// Методы для комментариев
	CreateCommentOnPost(ctx context.Context, comment *model.Comment) (*model.Comment, error)
	ReplyToComment(ctx context.Context, comment *model.Comment) (*model.Comment, error)
	GetReplies(parentID string) ([]*model.Comment, error)

	// // Получаем комментарии верхнего уровня для поста с пагинацией
	GetCommentsByPostID(postID string, limit, offset int) ([]*model.Comment, error)
}
