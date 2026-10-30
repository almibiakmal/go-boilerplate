package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// User adalah entity GORM untuk tabel users (mendukung soft delete via DeletedAt).
type User struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement"`
	Name      string         `gorm:"type:varchar(100);not null"`
	Email     string         `gorm:"type:varchar(150);uniqueIndex;not null"`
	Password  string         `gorm:"type:varchar(255);not null"`
	Role      string         `gorm:"type:varchar(32);not null;default:'user'"`
	CreatedAt time.Time      `gorm:"autoCreateTime"`
	UpdatedAt time.Time      `gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (User) TableName() string {
	return "users"
}
