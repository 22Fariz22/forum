package model

type User struct {
	ID       string `json:"id" gorm:"primaryKey;type:uuid"`
	Username string `json:"username" gorm:"type:varchar(20);not null"`
}

// Post – модель поста
type Post struct {
	ID            string     `json:"id" gorm:"primaryKey;type:uuid"`
	Title         string     `json:"title" gorm:"type:varchar(50);not null"`
	Content       string     `json:"content" gorm:"type:text;not null"`
	AllowComments bool       `json:"allowComments" gorm:"default:true"`
	AuthorID      string     `json:"authorID" gorm:"type:uuid;not null"`
	HaveComments  bool       `json:"haveComments" gorm:"default:false"`
	Comments      []*Comment `json:"comments" gorm:"foreignKey:PostID"`
}

// Comment – модель комментария
type Comment struct {
	ID           string  `json:"id" gorm:"primaryKey;type:uuid"`
	PostID       string  `json:"postID" gorm:"type:uuid;not null"`
	ParentID     *string `json:"parentID" gorm:"type:uuid"` // Если nil, то комментарий верхнего уровня
	Content      string  `json:"content" gorm:"type:text;not null"`
	Author       *User   `json:"author" gorm:"-"`
	HaveComments bool    `json:"haveComments" gorm:"default:false"`
}
