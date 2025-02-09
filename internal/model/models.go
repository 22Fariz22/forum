package model

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// Post – модель поста
type Post struct {
	ID            string     `json:"id"`
	Title         string     `json:"title"`
	Content       string     `json:"content"`
	AllowComments bool       `json:"allowComments"`
	AuthorID      string     `json:"author"`
	Comments      []*Comment `json:"comments"`
}

// Comment – модель комментария
type Comment struct {
	ID       string     `json:"id"`
	PostID   string     `json:"postID"`
	ParentID *string    `json:"parentID"` // Если nil, то комментарий верхнего уровня
	Content  string     `json:"content"`
	Author   *User      `json:"author"`
	Children []*Comment `json:"children"`
}
