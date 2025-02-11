package model

import "time"

// import "time"

type User struct {
	ID        string    `json:"id" gorm:"primaryKey;type:uuid"`
	Username  string    `json:"username" gorm:"type:varchar(20);not null"`
	CreatedAt time.Time `json:"createdAt" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}

// Post – модель поста
type Post struct {
	ID            string     `json:"id" gorm:"primaryKey;type:uuid"`
	Title         string     `json:"title" gorm:"type:varchar(50);not null"`
	Content       string     `json:"content" gorm:"type:text;not null"`
	AllowComments bool       `json:"allowComments" gorm:"default:true"`
	AuthorID      string     `json:"authorID" gorm:"type:uuid;not null"`
	HaveComments  bool       `json:"haveComments" gorm:"default:false"`
	Comments      []*Comment `json:"comments" gorm:"-"`
	CreatedAt     time.Time  `json:"createdAt" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}

// Comment – модель комментария
type Comment struct {
	ID           string    `json:"id" db:"id" gorm:"primaryKey;type:uuid"`
	PostID       string    `json:"postID" db:"post_id" gorm:"type:uuid;not null"`
	ParentID     *string   `json:"parentID" db:"parent_id" gorm:"type:uuid"` // Если nil, то комментарий верхнего уровня
	Content      string    `json:"content" db:"content" gorm:"type:text;not null"`
	Author       *User     `json:"author" gorm:"-"`
	AuthorID     string    `json:"authorID" db:"author_id" gorm:"foreignKey:AuthorID;references:ID"`
	Username     string    `json:"username" db:"username" gorm:"type:varchar(20);not null"`
	HaveComments bool      `json:"haveComments" db:"have_comments" gorm:"default:false"`
	CreatedAt    time.Time `json:"createdAt" gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
}
