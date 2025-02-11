package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.64

import (
	"context"
	"fmt"

	graphModel "github.com/22Fariz22/forum/graph/model"
	commonModel "github.com/22Fariz22/forum/internal/model"
	"github.com/22Fariz22/forum/utils"
	"github.com/google/uuid"
)

// CreatePost is the resolver for the createPost field.
func (r *mutationResolver) CreatePost(ctx context.Context, title string, content string, allowComments bool, author string) (*graphModel.Post, error) {
	// Проверяем, существует ли пользователь с таким ID
	_, err := r.Repo.GetUserByID(author)
	if err != nil {
		return nil, utils.NewGraphQLError("пользователь не найден", "404")
	}

	newPost := &commonModel.Post{
		ID:            uuid.New().String(),
		Title:         title,
		Content:       content,
		AllowComments: allowComments,
		AuthorID:      author,
	}

	newPostQLModel := &graphModel.Post{
		ID:            newPost.ID,
		Title:         newPost.Title,
		Content:       newPost.Content,
		AllowComments: newPost.AllowComments,
		AuthorID:      author,
	}

	//сохраняем в базе
	err = r.Repo.CreatePost(newPost)
	if err != nil {
		return nil, utils.NewGraphQLError("ошибка на сервере", "500")
	}

	return newPostQLModel, nil
}

// CreateCommentOnPost создаёт комментарий к посту
func (r *mutationResolver) CreateCommentOnPost(ctx context.Context, postID string, content string, author string) (*graphModel.Comment, error) {
	// Проверяем, существует ли пользователь с таким ID
	user, err := r.Repo.GetUserByID(author)
	if err != nil {
		return nil, utils.NewGraphQLError("пользователь не найден", "404")

	}

	// Проверяем, существует ли пост
	_, err = r.Repo.GetPostByID(postID)
	if err != nil {
		return nil, err
	}

	// Создаём комментарий
	comment := &commonModel.Comment{
		ID:      uuid.New().String(),
		PostID:  postID,
		Content: content,
		Author:  user,
	}

	// Добавляем комментарий
	c, err := r.Repo.CreateCommentOnPost(context.Background(), comment)
	if err != nil {
		fmt.Println("err in resolver call on r.Repo.CreateCommentOnPost: ", err)
		return nil, utils.NewGraphQLError("ошибка на сервере", "500")
	}

	return &graphModel.Comment{
		ID:        c.ID,
		PostID:    c.PostID,
		Content:   c.Content,
		Author:    (*graphModel.User)(c.Author),
		CreatedAt: c.CreatedAt,
	}, nil
}

// ReplyToComment создаёт ответ на комментарий
func (r *mutationResolver) ReplyToComment(ctx context.Context, parentID string, content string, author string) (*graphModel.Comment, error) {
	// Проверяем, существует ли пользователь с таким ID
	user, err := r.Repo.GetUserByID(author)
	if err != nil {
		return nil, utils.NewGraphQLError("пользователь не найден", "404")
	}

	// Создаём вложенный комментарий
	comment := &commonModel.Comment{
		ID:       uuid.New().String(),
		ParentID: &parentID,
		Content:  content,
		Author:   user,
	}

	// Добавляем комментарий
	c, err := r.Repo.ReplyToComment(context.Background(), comment)
	if err != nil {
		return nil, utils.NewGraphQLError("родительский коментарий не найден", "404")
	}

	return &graphModel.Comment{
		ID:      c.ID,
		PostID:  c.PostID,
		Content: c.Content,
		Author:  (*graphModel.User)(c.Author),
	}, nil
}

// CreateUser is the resolver for the createUser field.
func (r *mutationResolver) CreateUser(ctx context.Context, username string) (*graphModel.User, error) {
	user := &commonModel.User{
		ID:       uuid.New().String(),
		Username: username,
	}

	if len(username) < 2 {
		return nil, utils.NewGraphQLError("username должен быть длиной не меньше 2 символов", "400")
	}

	// Создаем пользователя
	if err := r.Repo.CreateUser(user); err != nil {
		fmt.Println("err:", err)
		return nil, utils.NewGraphQLError("ошибка на сервере", "500")
	}

	userGraphQL := graphModel.User{
		ID:       user.ID,
		Username: user.Username,
	}

	return &userGraphQL, nil
}

// Posts is the resolver for the posts field.
func (r *queryResolver) Posts(ctx context.Context) ([]*graphModel.Post, error) {
	// Получаем посты из репозитория
	posts, err := r.Repo.GetPosts()
	if err != nil {
		return nil, utils.NewGraphQLError("ошибка на сервере", "500")
	}

	// Преобразуем в GraphQL-модель и добавим в список
	var postsGraphQL []*graphModel.Post
	for _, post := range posts {
		postsGraphQL = append(postsGraphQL, &graphModel.Post{
			ID:            post.ID,
			Title:         post.Title,
			Content:       post.Content,
			AllowComments: post.AllowComments,
			HaveComments:  post.HaveComments,
			AuthorID:      post.AuthorID,
			CreatedAt:     post.CreatedAt,
		})
	}

	return postsGraphQL, nil
}

// Post is the resolver for the post field.
func (r *queryResolver) Post(ctx context.Context, id string) (*graphModel.Post, error) {
	// Получаем пост
	post, err := r.Repo.GetPostByID(id)
	if err != nil {
		return nil, utils.NewGraphQLError("пост не найден", "404")
	}

	// Загружаем комментарии с пагинацией
	comments, err := r.Repo.GetCommentsByPostID(post.ID, 10, 0)
	if err != nil {
		return nil, utils.NewGraphQLError("ошибка на сервере", "500")
	}

	// Преобразуем комментарии в graphModel и добавим в список
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
			HaveComments: comment.HaveComments,
			CreatedAt:    comment.CreatedAt,
		})
	}

	return &graphModel.Post{
		ID:            post.ID,
		Title:         post.Title,
		Content:       post.Content,
		AllowComments: post.AllowComments,
		AuthorID:      post.AuthorID,
		Comments:      graphComments,
		CreatedAt:     post.CreatedAt,
	}, nil
}

// GetReplies возвращает вложенные комментарии
func (r *queryResolver) GetReplies(ctx context.Context, parentID string) ([]*graphModel.Comment, error) {
	// Получаем список вложенных комментариев
	replies, err := r.Repo.GetReplies(parentID)
	if err != nil {
		return nil, utils.NewGraphQLError("ошибка на сервере", "500")
	}

	// Преобразуем их в GraphQL-модель
	var gqlReplies []*graphModel.Comment
	for _, c := range replies {
		gqlReplies = append(gqlReplies, &graphModel.Comment{
			ID:        c.ID,
			PostID:    c.PostID,
			ParentID:  c.ParentID,
			Content:   c.Content,
			CreatedAt: c.CreatedAt,
			Author:    (*graphModel.User)(c.Author),
		})
	}

	return gqlReplies, nil
}

// CommentAdded is the resolver for the commentAdded field.
func (r *subscriptionResolver) CommentAdded(ctx context.Context, postID string) (<-chan *graphModel.Comment, error) {
	return nil, nil
}

// Mutation returns MutationResolver implementation.
func (r *Resolver) Mutation() MutationResolver { return &mutationResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

// Subscription returns SubscriptionResolver implementation.
func (r *Resolver) Subscription() SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }
