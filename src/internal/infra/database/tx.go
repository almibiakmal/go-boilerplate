package database

import (
	"context"

	"gorm.io/gorm"
)

type txCtxKey struct{}

// TxManager menyediakan abstraksi untuk menjalankan satu atau beberapa operasi repository
// di dalam satu transaksi database atomik tanpa membocorkan *gorm.DB ke layer Service.
type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
}

type gormTxManager struct {
	db *gorm.DB
}

// NewTxManager membuat instance TxManager berbasis *gorm.DB.
func NewTxManager(db *gorm.DB) TxManager {
	return &gormTxManager{db: db}
}

// WithinTransaction membuka transaksi database, menyimpannya ke dalam txCtx,
// dan otomatis melakukan Commit jika fn mengembalikan nil, atau Rollback jika fn mengembalikan error/panic.
func (m *gormTxManager) WithinTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txCtxKey{}, tx)
		return fn(txCtx)
	})
}

// FromContext mengambil *gorm.DB aktif dari context jika dipanggil di dalam WithinTransaction,
// atau mengembalikan fallbackDB.WithContext(ctx) jika dipanggil di luar transaksi.
func FromContext(ctx context.Context, fallbackDB *gorm.DB) *gorm.DB {
	if ctx != nil {
		if tx, ok := ctx.Value(txCtxKey{}).(*gorm.DB); ok && tx != nil {
			return tx.WithContext(ctx)
		}
	}
	return fallbackDB.WithContext(ctx)
}

