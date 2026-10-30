package model

import "time"

// Product adalah entity database murni (GORM) untuk tabel products.
// Tidak menyertakan tag json agar tidak bocor langsung ke layer transport (HTTP/Event).
type Product struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement"`
	SKU       string    `gorm:"type:varchar(64);uniqueIndex;not null"`
	Name      string    `gorm:"type:varchar(150);not null"`
	Price     int64     `gorm:"not null;default:0"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}

func (Product) TableName() string {
	return "products"
}
