package repository

import "github.com/22Fariz22/forum/internal/model"

// Repository – интерфейс для работы с постами и комментариями
type Repository interface {
	// Методы для постов
	CreatePost(post *model.Post) error
	GetPosts() ([]*model.Post, error)
	GetPostByID(id string) (*model.Post, error)

	// Методы для комментариев
	CreateComment(comment *model.Comment) error
	// Получаем комментарии верхнего уровня для поста с пагинацией
	GetCommentsByPostID(postID string, limit, offset int) ([]*model.Comment, error)
	// Получаем вложенные комментарии для родительского комментария с пагинацией
	GetCommentsByParentID(parentID string, limit, offset int) ([]*model.Comment, error)
}
