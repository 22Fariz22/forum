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
	users         map[string]*model.User
	posts         map[string]*model.Post      //посты по post_id(выдаем при просмотре одного поста за O(1))
	sortedPosts   []*model.Post               //отсортированные посты по CreatedAt(выдача всех постов за О(1))
	comments      map[string][]*model.Comment //key=post_id
	replyComments map[string][]*model.Comment //key=parentID
	subscribers   map[string][]chan *model.Comment
	mu            sync.RWMutex
}

func NewInMemoryRepository() Repository {
	return &InMemoryRepository{
		users:         make(map[string]*model.User),
		posts:         make(map[string]*model.Post),
		sortedPosts:   []*model.Post{},
		comments:      make(map[string][]*model.Comment),
		replyComments: make(map[string][]*model.Comment),
		subscribers:   make(map[string][]chan *model.Comment),
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
		fmt.Println("if !exists", exists)
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

	// Добавляем в slice
	r.sortedPosts = append(r.sortedPosts, post)

	// Вызываем сортировку
	r.sortPosts()

	// Создаём пустой список комментариев для поста
	r.comments[post.ID] = []*model.Comment{}

	return nil
}

// sortPosts сортирует r.sortedPosts по CreatedAt (новые сверху)
func (r *InMemoryRepository) sortPosts() {
	sort.Slice(r.sortedPosts, func(i, j int) bool {
		return r.sortedPosts[i].CreatedAt.After(r.sortedPosts[j].CreatedAt)
	})
}

// GetPosts возвращает все посты
func (r *InMemoryRepository) GetPosts(offset int32, limit int32) ([]*model.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Проверяем границы пагинации
	if offset < 0 || limit <= 0 {
		return nil, errors.New("неверные параметры пагинации")
	}

	// Ограничиваем список постов
	start := int(offset)
	end := start + int(limit)
	if end > len(r.sortedPosts) {
		end = len(r.sortedPosts)
	}

	// Возвращаем срез постов
	return r.sortedPosts[offset:end], nil
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

	// Проверяем, является ли комментарий верхнеуровневым
	if comment.ParentID == nil {
		// Инициализируем массив комментариев для поста, если его нет
		if _, ok := r.comments[comment.PostID]; !ok {
			r.comments[comment.PostID] = []*model.Comment{}
		}

		// Добавляем комментарий
		comment.CreatedAt = time.Now()
		r.comments[comment.PostID] = append(r.comments[comment.PostID], comment)

		// Инициализируем массив для вложенных комментариев
		r.replyComments[comment.ID] = []*model.Comment{}

		return comment, nil
	}

	// Проверяем, существует ли родительский комментарий
	parentID := *comment.ParentID

	replies, ok := r.replyComments[parentID]
	if !ok {
		fmt.Println("Родительский комментарий не найден:", parentID)
		return nil, errors.New("родительский комментарий не найден")
	}

	// Добавляем время создания
	comment.CreatedAt = time.Now()

	// Добавляем комментарий в список ответов
	r.replyComments[parentID] = append(replies, comment)

	// Сортируем вложенные комментарии по времени (новые сверху)
	sortCommentsByCreatedAt(r.replyComments[parentID])

	return comment, nil
}

// sortCommentsByCreatedAt сортирует комментарии по времени (новые сверху)
func sortCommentsByCreatedAt(comments []*model.Comment) {
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.After(comments[j].CreatedAt)
	})
}

// ReplyToComment добавляет ответ на комментарий
func (r *InMemoryRepository) ReplyToComment(ctx context.Context, comment *model.Comment) (*model.Comment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Проверяем, указан ли parentID
	if comment.ParentID == nil {
		return nil, errors.New("parentID не может быть nil")
	}

	parentID := *comment.ParentID

	// Проверяем, существует ли родительский комментарий в `replyComments`
	if _, exists := r.replyComments[parentID]; !exists {
		return nil, errors.New("родительский комментарий не найден")
	}

	// Добавляем время создания
	comment.CreatedAt = time.Now()

	// Добавляем комментарий в список ответов
	r.replyComments[parentID] = append(r.replyComments[parentID], comment)

	// Сортируем вложенные комментарии по времени (новые сверху)
	sortCommentsByCreatedAt(r.replyComments[parentID])

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
func (r *InMemoryRepository) GetCommentsByPostID(postID string, offset, limit int) ([]*model.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Проверяем, есть ли комментарии у поста
	comments, exists := r.comments[postID]
	if !exists || len(comments) == 0 {
		fmt.Println("if !exists || len(comments) == 0 ", !exists, len(comments) == 0)
		return []*model.Comment{}, nil
	}

	// Проверяем, что offset не выходит за границы
	if offset < 0 || offset >= len(comments) {
		return []*model.Comment{}, nil
	}

	end := offset + limit
	if end > len(comments) {
		fmt.Println("		end = len(comments)")
		end = len(comments)
	}

	return comments[offset:end], nil
}

// GetReplies возвращает вложенные комментарии по parentID
func (r *InMemoryRepository) GetReplies(parentID string, offset, limit int) ([]*model.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	replies, exists := r.replyComments[parentID]
	if !exists || len(replies) == 0 {
		fmt.Println("	Нет вложенных комментариев для parentID:", parentID)
		return []*model.Comment{}, nil
	}

	// Проверяем, что offset не выходит за границы
	if offset < 0 || offset >= len(replies) {
		fmt.Println("		Offset выходит за пределы списка. Offset:", offset, "Length of replies:", len(replies))
		return []*model.Comment{}, nil
	}

	// Определяем границы среза с учётом лимита
	end := offset + limit
	if end > len(replies) {
		fmt.Println("	Конечный индекс превышает длину списка, устанавливаем end:", len(replies))
		end = len(replies)
	}

	return replies[offset:end], nil
}
