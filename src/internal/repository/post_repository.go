package repository

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	servicedto "maj-core-srvc/src/internal/dto/service"
	"maj-core-srvc/src/internal/infra/database"
	"maj-core-srvc/src/internal/model"
)

type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	FindByID(ctx context.Context, id uint64) (*model.Post, error)
	List(ctx context.Context, filter servicedto.ListPostFilter) ([]model.Post, int64, error)
	Update(ctx context.Context, post *model.Post) error
	Delete(ctx context.Context, id uint64, authorID uint64) (bool, error)
	AdjustCommentsCount(ctx context.Context, postID uint64, delta int) error
}

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post *model.Post) error {
	return database.FromContext(ctx, r.db).Create(post).Error
}

func (r *postRepository) FindByID(ctx context.Context, id uint64) (*model.Post, error) {
	var post model.Post
	err := database.FromContext(ctx, r.db).Preload("Author").First(&post, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) List(ctx context.Context, f servicedto.ListPostFilter) ([]model.Post, int64, error) {
	var total int64

	tx := database.FromContext(ctx, r.db).Model(&model.Post{})
	if f.AuthorID > 0 {
		tx = tx.Where("author_id = ?", f.AuthorID)
	}
	if f.Search != "" {
		like := "%" + strings.ToLower(f.Search) + "%"
		tx = tx.Where("LOWER(title) LIKE ? OR LOWER(content) LIKE ?", like, like)
	}

	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []model.Post{}, 0, nil
	}

	var posts []model.Post
	offset := (f.Page - 1) * f.PerPage
	if err := tx.Preload("Author").Order("id DESC").Limit(f.PerPage).Offset(offset).Find(&posts).Error; err != nil {
		return nil, 0, err
	}

	return posts, total, nil
}

func (r *postRepository) Update(ctx context.Context, post *model.Post) error {
	res := database.FromContext(ctx, r.db).
		Model(&model.Post{}).
		Where("id = ? AND author_id = ?", post.ID, post.AuthorID).
		Updates(map[string]any{
			"title":        post.Title,
			"content":      post.Content,
			"is_published": post.IsPublished,
		})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *postRepository) Delete(ctx context.Context, id uint64, authorID uint64) (bool, error) {
	res := database.FromContext(ctx, r.db).
		Where("id = ? AND author_id = ?", id, authorID).
		Delete(&model.Post{})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

// AdjustCommentsCount mengubah counter comments_count secara atomik di dalam transaksi (ANSI SQL compatible untuk Postgres, MySQL, SQLite).
func (r *postRepository) AdjustCommentsCount(ctx context.Context, postID uint64, delta int) error {
	return database.FromContext(ctx, r.db).
		Model(&model.Post{}).
		Where("id = ?", postID).
		UpdateColumn("comments_count", gorm.Expr("CASE WHEN comments_count + ? < 0 THEN 0 ELSE comments_count + ? END", delta, delta)).
		Error
}
