package model

import (
	"time"

	"gorm.io/gorm"
)

// Post adalah entity GORM untuk tabel posts (mendukung soft delete via DeletedAt).
type Post struct {
	ID            uint64         `gorm:"primaryKey;autoIncrement"`
	AuthorID      uint64         `gorm:"not null;index"`
	Author        *User          `gorm:"foreignKey:AuthorID"`
	Title         string         `gorm:"type:varchar(200);not null"`
	Slug          string         `gorm:"type:varchar(220);uniqueIndex;not null"`
	Content       string         `gorm:"type:text;not null"`
	IsPublished   bool           `gorm:"column:is_published;not null;default:true"`
	CommentsCount int            `gorm:"not null;default:0"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (Post) TableName() string {
	return "posts"
}
