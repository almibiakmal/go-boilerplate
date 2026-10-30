package model

import (
	"time"

	"gorm.io/gorm"
)

// Comment adalah entity GORM untuk tabel comments (mendukung soft delete via DeletedAt).
type Comment struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement"`
	PostID    uint64         `gorm:"not null;index"`
	AuthorID  uint64         `gorm:"not null;index"`
	Author    *User          `gorm:"foreignKey:AuthorID"`
	Content   string         `gorm:"type:text;not null"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Comment) TableName() string {
	return "comments"
}
