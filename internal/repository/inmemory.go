package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/22Fariz22/forum/internal/model"
)

type InMemoryRepository struct {
	users       map[string]*model.User
	posts       map[string]*model.Post // храним посты по ID(выдаем при просмотре одного поста за O(1))
	sortedPosts []*model.Post          // Храним посты отсортированными по CreatedAt(при просмотре всех постов за О(1))
	Comments    map[string]*model.Comment
	subscribers map[string][]chan *model.Comment
	mu          sync.RWMutex
}

func NewInMemoryRepository() Repository {
	return &InMemoryRepository{
		users:       make(map[string]*model.User),
		posts:       make(map[string]*model.Post),
		Comments:    make(map[string]*model.Comment),
		subscribers: make(map[string][]chan *model.Comment),
	}
}

// CreateUser добавляет нового пользователя, если его еще нет
func (r *InMemoryRepository) CreateUser(user *model.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.users[user.ID]; exists {
		return errors.New("user already exists")
	}

	r.users[user.ID] = user
	return nil
}

// GetUserByID проверяет существование пользователя
func (r *InMemoryRepository) GetUserByID(id string) (*model.User, error) {
	r.mu.RLock()
	defer func() {
		r.mu.RUnlock()
	}()

	user, exists := r.users[id]

	if !exists {
		return nil, errors.New("user not found")
	}

	return user, nil
}

// CreatePost добавляет новый пост, если автор существует
func (r *InMemoryRepository) CreatePost(post *model.Post) error {
	// Проверяем существование пользователя перед созданием поста
	_, err := r.GetUserByID(post.AuthorID)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	post.CreatedAt = time.Now()
	r.posts[post.ID] = post

	// Добавляем в slice и сортируем
	r.sortedPosts = append(r.sortedPosts, post)

	// Вызываем сортировку
	r.sortPosts()

	return nil
}

// sortPosts сортирует r.sortedPosts по CreatedAt (новые сверху)
func (r *InMemoryRepository) sortPosts() {
	sort.Slice(r.sortedPosts, func(i, j int) bool {
		return r.sortedPosts[i].CreatedAt.After(r.sortedPosts[j].CreatedAt)
	})
}

// GetPosts возвращает все посты
func (r *InMemoryRepository) GetPosts() ([]*model.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.sortedPosts, nil
}

// GetPostByID возвращает пост по ID
func (r *InMemoryRepository) GetPostByID(id string) (*model.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if post, exists := r.posts[id]; exists {
		return post, nil
	}

	return nil, errors.New("пост не найден")
}

// CreateCommentOnPost добавляет комментарий к посту
func (r *InMemoryRepository) CreateCommentOnPost(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем, существует ли пост
	if _, exists := r.posts[comment.PostID]; !exists {
		return nil, errors.New("пост не найден")
	}

	// Добавляем комментарий
	r.Comments[comment.ID] = comment

	// Обновляем флаг наличия комментариев у поста
	r.posts[comment.PostID].HaveComments = true

	//добавляем время создания
	comment.CreatedAt = time.Now()

	// Уведомляем подписчиков
	go r.NotifySubscribers(comment.PostID, comment)

	return comment, nil
}

// ReplyToComment добавляет ответ на комментарий
func (r *InMemoryRepository) ReplyToComment(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем, существует ли родительский комментарий
	parentComment, exists := r.Comments[*comment.ParentID]
	if !exists {
		return nil, errors.New("родительский комментарий не найден")
	}

	// Добавляем комментарий
	r.Comments[comment.ID] = comment

	// Устанавливаем флаг, что у родительского комментария есть ответы
	parentComment.HaveComments = true

	return comment, nil
}

// NotifySubscribers отправляет новый комментарий подписчикам
func (r *InMemoryRepository) NotifySubscribers(postID string, comment *model.Comment) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if chans, exists := r.subscribers[postID]; exists {
		for _, ch := range chans {
			select {
			case ch <- comment: // Отправляем комментарий в канал
			default:
				fmt.Println("Канал подписчика переполнен, пропускаем уведомление")
			}
		}
	}
}

func (r *InMemoryRepository) SubscribeToComments(postID string) <-chan *model.Comment {
	r.mu.Lock()
	defer r.mu.Unlock()

	ch := make(chan *model.Comment, 1)
	r.subscribers[postID] = append(r.subscribers[postID], ch)
	return ch
}

// GetCommentsByPostID получает комментарии верхнего уровня
func (r *InMemoryRepository) GetCommentsByPostID(postID string, limit, offset int) ([]*model.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var comments []*model.Comment
	for _, comment := range r.Comments {
		if comment.PostID == postID && comment.ParentID == nil {
			comments = append(comments, comment)
		}
	}

	// Пагинация
	if offset >= len(comments) {
		return []*model.Comment{}, nil
	}
	end := offset + limit
	if end > len(comments) {
		end = len(comments)
	}
	return comments[offset:end], nil
}

// GetReplies возвращает вложенные комментарии по parentID
func (r *InMemoryRepository) GetReplies(parentID string) ([]*model.Comment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var replies []*model.Comment
	for _, comment := range r.Comments {
		if comment.ParentID != nil && *comment.ParentID == parentID {
			replies = append(replies, comment)
		}
	}

	return replies, nil
}
