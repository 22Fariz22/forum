package graph

import (
	"context"
	"errors"
	"fmt"

	"github.com/22Fariz22/forum/internal/model"
	graphModel "github.com/22Fariz22/forum/internal/model"
	"github.com/22Fariz22/forum/pubsub"
	"github.com/22Fariz22/forum/repository"
	"github.com/google/uuid"
)

// Repository определён в пакете repository
type Repository = repository.Repository

// Resolver содержит ссылки на хранилище и систему pubsub для подписок.
type Resolver struct {
	Repo   Repository
	PubSub *pubsub.PubSub
}

func NewResolver(repo Repository) *Resolver {
	return &Resolver{
		Repo:   repo,
		PubSub: pubsub.NewPubSub(),
	}
}

// ------------------
// РЕЗОЛВЕРЫ ЗАПРОСОВ (Query)
// ------------------

func (r *Resolver) Posts(ctx context.Context) ([]*model.Post, error) {
	return r.Repo.GetPosts()
}

func (r *Resolver) Post(ctx context.Context, id string) (*model.Post, error) {
	return r.Repo.GetPostByID(id)
}

// ------------------
// РЕЗОЛВЕРЫ МУТАЦИЙ (Mutation)
// ------------------

// CreatePost создает новый пост с указанием автора (передается как строка username)
func (r *Resolver) CreatePost(ctx context.Context, title string, content string, allowComments bool, author string) (*model.Post, error) {
	// Создаем объект User для автора поста
	fmt.Println("CREATEPOST in resolver.go")
	user := &model.User{
		ID:       uuid.New().String(),
		Username: author,
	}

	newPost := &model.Post{
		ID:            uuid.New().String(),
		Title:         title,
		Content:       content,
		AllowComments: allowComments,
		Author:        user,
	}

	err := r.Repo.CreatePost(newPost)
	if err != nil {
		return nil, err
	}

	newPostQLModel := &graphModel.Post{
		ID:            newPost.ID,
		Title:         newPost.Title,
		Content:       newPost.Content,
		AllowComments: newPost.AllowComments,
		Author:        (*graphModel.User)(newPost.Author),
	}

	return newPostQLModel, nil
}

// CreateComment создает комментарий с указанием автора
func (r *Resolver) CreateComment(ctx context.Context, postID string, parentID *string, content string, author string) (*model.Comment, error) {
	// Проверяем, что пост существует
	post, err := r.Repo.GetPostByID(postID)
	if err != nil {
		return nil, errors.New("пост не найден")
	}
	// Если комментарии запрещены
	if !post.AllowComments {
		return nil, errors.New("комментарии запрещены для этого поста")
	}
	// Ограничиваем длину текста комментария (например, до 2000 символов)
	if len(content) > 2000 {
		return nil, errors.New("длина комментария превышает лимит")
	}

	// Создаем объект User для автора комментария
	user := &model.User{
		ID:       uuid.New().String(),
		Username: author,
	}

	newComment := &model.Comment{
		ID:       uuid.New().String(),
		PostID:   postID,
		ParentID: parentID,
		Content:  content,
		Author:   user,
	}

	err = r.Repo.CreateComment(newComment)
	if err != nil {
		return nil, err
	}

	// Публикуем уведомление для подписанных клиентов
	topic := fmt.Sprintf("post_%s", postID)
	r.PubSub.Publish(topic, newComment)
	return newComment, nil
}

// ------------------
// РЕЗОЛВЕР ПОДПИСОК (Subscription)
// ------------------

func (r *Resolver) CommentAdded(ctx context.Context, postID string) (<-chan *model.Comment, error) {
	topic := fmt.Sprintf("post_%s", postID)
	sub := r.PubSub.Subscribe(topic)

	out := make(chan *model.Comment)
	go func() {
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				r.PubSub.Unsubscribe(topic, sub)
				return
			case msg, ok := <-sub:
				if !ok {
					return
				}
				if comment, ok := msg.(*model.Comment); ok {
					out <- comment
				}
			}
		}
	}()
	return out, nil
}
