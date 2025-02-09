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

// CreatePost is the resolver for the createPost field.
func (r *mutationResolver) CreatePost(ctx context.Context, title string, content string, allowComments bool, author string) (*graphModel.Post, error) {
	// Создаем объект User для автора поста
	fmt.Println("in resolver CreatePost()")

	// Проверяем, существует ли пользователь с таким ID
	if err := r.Repo.GetUserByID(author); err != nil {
		return nil, errors.New("пользователь не найден")
	}

	newPost := &commonModel.Post{
		ID:            uuid.New().String(),
		Title:         title,
		Content:       content,
		AllowComments: allowComments,
		AuthorID:      author,
	}

	//сохраняем в базе
	err := r.Repo.CreatePost(newPost)
	if err != nil {
		return nil, err
	}

	newPostQLModel := &graphModel.Post{
		ID:            newPost.ID,
		Title:         newPost.Title,
		Content:       newPost.Content,
		AllowComments: newPost.AllowComments,
		AuthorID:      author,
	}

	return newPostQLModel, nil
}

// CreateComment
func (r *mutationResolver) CreateComment(ctx context.Context, postID string, parentID *string, content string, author string) (*graphModel.Comment, error) {
	panic(fmt.Errorf("not implemented: CreateComment - createComment"))
}

// CreateUser
func (r *mutationResolver) CreateUser(ctx context.Context, username string) (*graphModel.User, error) {
	fmt.Println("in resolver CreateUser()")

	user := &commonModel.User{
		ID:       uuid.New().String(),
		Username: username,
	}

	// Создаем пользователя
	if err := r.Repo.CreateUser(user); err != nil {
		return nil, err
	}

	userGraphQL := graphModel.User{
		ID:       user.ID,
		Username: user.Username,
	}

	return &userGraphQL, nil
}

// Posts is the resolver for the posts field.
func (r *queryResolver) Posts(ctx context.Context) ([]*graphModel.Post, error) {
	fmt.Println("Получаем все посты...")

	// Получаем посты из репозитория
	posts, err := r.Repo.GetPosts()
	if err != nil {
		return nil, err
	}

	// Преобразуем в GraphQL-модель
	var postsGraphQL []*graphModel.Post
	for _, post := range posts {
		postsGraphQL = append(postsGraphQL, &graphModel.Post{
			ID:            post.ID,
			Title:         post.Title,
			Content:       post.Content,
			AllowComments: post.AllowComments,
			AuthorID:      post.AuthorID,
		})
	}

	return postsGraphQL, nil
}

// Post is the resolver for the post field.
func (r *queryResolver) Post(ctx context.Context, id string) (*graphModel.Post, error) {
	fmt.Println("получаем пост ID:", id)

	// Получаем пост
	post, err := r.Repo.GetPostByID(id)
	if err != nil {
		return nil, err
	}

	// Загружаем комментарии с пагинацией
	comments, err := r.Repo.GetCommentsByPostID(post.ID, 10, 0)
	if err != nil {
		return nil, err
	}
	fmt.Println("comments in resolver:")
	for i, v := range comments {
		fmt.Println(i, v.Content, v.Author.Username)
	}

	// Преобразуем комментарии в graphModel
	var graphComments []*graphModel.Comment
	for _, comment := range comments {
		graphComments = append(graphComments, &graphModel.Comment{
			ID:       comment.ID,
			PostID:   comment.PostID,
			ParentID: comment.ParentID,
			Content:  comment.Content,
			Author: &graphModel.User{
				ID:       comment.Author.ID,
				Username: comment.Author.Username,
			},
		})
	}

	return &graphModel.Post{
		ID:            post.ID,
		Title:         post.Title,
		Content:       post.Content,
		AllowComments: post.AllowComments,
		AuthorID:      post.AuthorID,
		Comments:      graphComments,
	}, nil
}

// CommentAdded is the resolver for the commentAdded field.
func (r *subscriptionResolver) CommentAdded(
	ctx context.Context,
	postID string,
) (<-chan *graphModel.Comment, error) {

	// Подписываемся на новые комментарии в репозитории
	commentCh := r.Repo.(*repository.InMemoryRepository).SubscribeToComments(postID)

	// Создаём новый канал для преобразования типа
	outCh := make(chan *graphModel.Comment, 1)

	go func() {
		for comment := range commentCh {
			outCh <- &graphModel.Comment{
				ID:       comment.ID,
				PostID:   comment.PostID,
				ParentID: comment.ParentID,
				Content:  comment.Content,
				Author: &graphModel.User{
					ID:       comment.Author.ID,
					Username: comment.Author.Username,
				},
			}
		}
		close(outCh) // Закрываем канал, когда подписка завершена
	}()

	return outCh, nil

}

/*
schema {
  query: Query
  mutation: Mutation
  subscription: Subscription
}

type User {
  id: ID!
  username: String!
}

type Post {
  id: ID!
  title: String!
  content: String!
  allowComments: Boolean!
  authorID: ID!
  # Пагинация комментариев (только верхнего уровня)
  comments(limit: Int = 10, offset: Int = 0): [Comment!]!
}

type Comment {
  id: ID!
  postID: ID!
  parentID: ID
  content: String!
  author: User!
  # Для вложенных комментариев
  children(limit: Int = 10, offset: Int = 0): [Comment!]!
}

type Query {
  posts: [Post!]!
  post(id: ID!): Post
}

type Mutation {
  createPost(
    title: String!
    content: String!
    allowComments: Boolean!
    author: ID!
  ): Post!
  createComment(
    postID: ID!
    parentID: ID
    content: String!
    author: ID!
  ): Comment!
  createUser(username: String!): User!
}

type Subscription {
  commentAdded(postID: ID!): Comment!
}
*/
