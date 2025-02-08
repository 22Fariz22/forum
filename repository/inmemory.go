package repository

import (
	"errors"
	"sync"

	"github.com/22Fariz22/forum/internal/model"
)

type InMemoryRepository struct {
	posts    map[string]*model.Post
	comments map[string]*model.Comment
	mu       sync.RWMutex
}

func NewInMemoryRepository() Repository {
	return &InMemoryRepository{
		posts:    make(map[string]*model.Post),
		comments: make(map[string]*model.Comment),
	}
}

func (r *InMemoryRepository) CreatePost(post *model.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.posts[post.ID]; exists {
		return errors.New("пост уже существует")
	}
	r.posts[post.ID] = post
	return nil
}

func (r *InMemoryRepository) GetPosts() ([]*model.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var posts []*model.Post
	for _, post := range r.posts {
		posts = append(posts, post)
	}
	return posts, nil
}

func (r *InMemoryRepository) GetPostByID(id string) (*model.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if post, exists := r.posts[id]; exists {
		return post, nil
	}
	return nil, errors.New("пост не найден")
}

func (r *InMemoryRepository) CreateComment(comment *model.Comment) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.comments[comment.ID]; exists {
		return errors.New("комментарий уже существует")
	}
	r.comments[comment.ID] = comment
	return nil
}

func (r *InMemoryRepository) GetCommentsByPostID(postID string, limit, offset int) ([]*model.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var comments []*model.Comment
	for _, comment := range r.comments {
		// выбираем только комментарии верхнего уровня (без родительского комментария)
		if comment.PostID == postID && comment.ParentID == nil {
			comments = append(comments, comment)
		}
	}
	// Простейшая пагинация
	if offset > len(comments) {
		return []*model.Comment{}, nil
	}
	end := offset + limit
	if end > len(comments) {
		end = len(comments)
	}
	return comments[offset:end], nil
}

func (r *InMemoryRepository) GetCommentsByParentID(parentID string, limit, offset int) ([]*model.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var comments []*model.Comment
	for _, comment := range r.comments {
		if comment.ParentID != nil && *comment.ParentID == parentID {
			comments = append(comments, comment)
		}
	}
	if offset > len(comments) {
		return []*model.Comment{}, nil
	}
	end := offset + limit
	if end > len(comments) {
		end = len(comments)
	}
	return comments[offset:end], nil
}
