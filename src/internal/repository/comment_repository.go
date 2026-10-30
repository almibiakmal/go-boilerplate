package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"

	servicedto "maj-core-srvc/src/internal/dto/service"
	"maj-core-srvc/src/internal/infra/database"
	"maj-core-srvc/src/internal/model"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *model.Comment) error
	FindByID(ctx context.Context, id uint64) (*model.Comment, error)
	ListByPostID(ctx context.Context, filter servicedto.ListCommentFilter) ([]model.Comment, int64, error)
	Update(ctx context.Context, comment *model.Comment) error
	Delete(ctx context.Context, id uint64) (bool, error)
	DeleteByPostID(ctx context.Context, postID uint64) error
}

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *model.Comment) error {
	return database.FromContext(ctx, r.db).Create(comment).Error
}

func (r *commentRepository) FindByID(ctx context.Context, id uint64) (*model.Comment, error) {
	var comment model.Comment
	err := database.FromContext(ctx, r.db).Preload("Author").First(&comment, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *commentRepository) ListByPostID(ctx context.Context, f servicedto.ListCommentFilter) ([]model.Comment, int64, error) {
	var total int64

	tx := database.FromContext(ctx, r.db).Model(&model.Comment{}).Where("post_id = ?", f.PostID)
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if total == 0 {
		return []model.Comment{}, 0, nil
	}

	var comments []model.Comment
	offset := (f.Page - 1) * f.PerPage
	if err := tx.Preload("Author").Order("id ASC").Limit(f.PerPage).Offset(offset).Find(&comments).Error; err != nil {
		return nil, 0, err
	}

	return comments, total, nil
}

func (r *commentRepository) Update(ctx context.Context, comment *model.Comment) error {
	return database.FromContext(ctx, r.db).Save(comment).Error
}

func (r *commentRepository) Delete(ctx context.Context, id uint64) (bool, error) {
	res := database.FromContext(ctx, r.db).Delete(&model.Comment{}, id)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *commentRepository) DeleteByPostID(ctx context.Context, postID uint64) error {
	return database.FromContext(ctx, r.db).Where("post_id = ?", postID).Delete(&model.Comment{}).Error
}

