package repository

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/22Fariz22/forum/internal/model"
	"github.com/google/uuid"
)

type InMemoryRepository struct {
	users       map[string]*model.User
	posts       map[string]*model.Post
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
	fmt.Println("in InMemoryRepository CreateUser()")
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
	fmt.Println("in InMemoryRepository GetUserByID()", "id:", id)
	r.mu.RLock()
	defer func() {
		fmt.Println("unlocking GetUserByID()")
		r.mu.RUnlock()
	}()

	user, exists := r.users[id]

	if !exists {
		fmt.Println("user not found:", id)
		return nil, errors.New("user not found")
	}

	fmt.Println("user found:", user.ID, user.Username)
	return user, nil
}

// CreatePost добавляет новый пост, если автор существует
func (r *InMemoryRepository) CreatePost(post *model.Post) error {
	fmt.Println("in InMemoryRepository CreatePost()")

	// Проверяем существование пользователя перед созданием поста
	_, err := r.GetUserByID(post.AuthorID)
	if err != nil {
		fmt.Println("User not found, cannot create post:", post.AuthorID)
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.posts[post.ID] = post
	fmt.Println("Post created:", post.ID)
	return nil
}

// GetPosts возвращает все посты
func (r *InMemoryRepository) GetPosts() ([]*model.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	posts := make([]*model.Post, 0, len(r.posts))
	for _, post := range r.posts {
		posts = append(posts, post)
	}
	return posts, nil
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
	fmt.Println("call inmemory GetCommentsByPostID ")
	r.mu.RLock()
	defer r.mu.RUnlock()

	var comments []*model.Comment
	for _, comment := range r.Comments {
		if comment.PostID == postID && comment.ParentID == nil {
			fmt.Println("comment:", comment.Content)
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
	fmt.Println("ofset:", comments[offset:end])
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

func (r *InMemoryRepository) SeedData() {
	r.mu.Lock()
	defer r.mu.Unlock()

	fmt.Println("Seeding initial data...")

	// Создаём пользователей
	user1 := &model.User{ID: uuid.New().String(), Username: "Alice"}
	user2 := &model.User{ID: uuid.New().String(), Username: "Bob"}
	user3 := &model.User{ID: uuid.New().String(), Username: "Charlie"}

	r.users[user1.ID] = user1
	r.users[user2.ID] = user2
	r.users[user3.ID] = user3

	// Создаём пост
	post1 := &model.Post{
		ID:            uuid.New().String(),
		Title:         "Пост про собаку",
		Content:       "Собаку зовут Бобик",
		AllowComments: true,
		AuthorID:      user1.ID,
		Comments:      []*model.Comment{},
	}

	r.posts[post1.ID] = post1

	// Создаём комментарии
	comment1 := &model.Comment{
		ID:       uuid.New().String(),
		PostID:   post1.ID,
		ParentID: nil,
		Content:  "Какого цвета собака?",
		Author:   user2,
	}

	comment2 := &model.Comment{
		ID:       uuid.New().String(),
		PostID:   post1.ID,
		ParentID: nil,
		Content:  "Черный",
		Author:   user1,
	}

	// Вложенные комментарии к `comment2`
	comment5 := &model.Comment{
		ID:       uuid.New().String(),
		PostID:   post1.ID,
		ParentID: &comment2.ID,
		Content:  "Классный цвет!",
		Author:   user3,
	}

	comment6 := &model.Comment{
		ID:       uuid.New().String(),
		PostID:   post1.ID,
		ParentID: &comment2.ID,
		Content:  "Люблю черных собак!",
		Author:   user2,
	}

	// Добавляем комментарии в пост
	post1.Comments = append(post1.Comments, comment1, comment2)

	// Добавляем в общую структуру
	r.Comments[comment1.ID] = comment1
	r.Comments[comment2.ID] = comment2
	r.Comments[comment5.ID] = comment5
	r.Comments[comment6.ID] = comment6

	fmt.Println("Seeding completed.")
}
