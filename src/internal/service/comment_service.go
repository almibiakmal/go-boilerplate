package service

import (
	"context"
	"strings"

	responsedto "maj-core-srvc/src/internal/dto/response"
	servicedto "maj-core-srvc/src/internal/dto/service"
	"maj-core-srvc/src/internal/infra/database"
	"maj-core-srvc/src/internal/infra/logger"
	"maj-core-srvc/src/internal/model"
	"maj-core-srvc/src/internal/pkg/response"
	"maj-core-srvc/src/internal/pkg/validator"
	"maj-core-srvc/src/internal/repository"
)

type CommentService interface {
	Create(ctx context.Context, input servicedto.CreateCommentInput) (*responsedto.CommentResponse, error)
	ListByPostID(ctx context.Context, filter servicedto.ListCommentFilter) ([]responsedto.CommentResponse, int64, error)
	Update(ctx context.Context, input servicedto.UpdateCommentInput) (*responsedto.CommentResponse, error)
	Delete(ctx context.Context, input servicedto.DeleteCommentInput) error
}

type commentService struct {
	txManager   database.TxManager
	postRepo    repository.PostRepository
	commentRepo repository.CommentRepository
}

func NewCommentService(
	txManager database.TxManager,
	postRepo repository.PostRepository,
	commentRepo repository.CommentRepository,
) CommentService {
	return &commentService{
		txManager:   txManager,
		postRepo:    postRepo,
		commentRepo: commentRepo,
	}
}

func (s *commentService) Create(ctx context.Context, input servicedto.CreateCommentInput) (*responsedto.CommentResponse, error) {
	input.Content = strings.TrimSpace(input.Content)

	if err := validator.ValidateStruct(input); err != nil {
		return nil, err
	}

	var createdComment *model.Comment

	// Database Transaction: pastikan post ada, simpan komentar, dan tambah counter comments_count secara atomik
	err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		post, err := s.postRepo.FindByID(txCtx, input.PostID)
		if err != nil {
			return response.WrapError(response.CodeDatabaseError, "Failed to verify post", err)
		}
		if post == nil {
			return response.NewError(response.CodeNotFound, "Post not found")
		}

		comment := &model.Comment{
			PostID:   input.PostID,
			AuthorID: input.AuthorID,
			Content:  input.Content,
		}

		if err := s.commentRepo.Create(txCtx, comment); err != nil {
			return response.WrapError(response.CodeDatabaseError, "Failed to create comment", err)
		}

		if err := s.postRepo.AdjustCommentsCount(txCtx, input.PostID, 1); err != nil {
			return response.WrapError(response.CodeDatabaseError, "Failed to increment post comments count", err)
		}

		withAuthor, err := s.commentRepo.FindByID(txCtx, comment.ID)
		if err == nil && withAuthor != nil {
			createdComment = withAuthor
		} else {
			createdComment = comment
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "comment created", "comment_id", createdComment.ID, "post_id", input.PostID, "author_id", input.AuthorID)
	return responsedto.ToCommentResponse(createdComment), nil
}

func (s *commentService) ListByPostID(ctx context.Context, filter servicedto.ListCommentFilter) ([]responsedto.CommentResponse, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 10
	}

	post, err := s.postRepo.FindByID(ctx, filter.PostID)
	if err != nil {
		return nil, 0, response.WrapError(response.CodeDatabaseError, "Failed to verify post", err)
	}
	if post == nil {
		return nil, 0, response.NewError(response.CodeNotFound, "Post not found")
	}

	items, total, err := s.commentRepo.ListByPostID(ctx, filter)
	if err != nil {
		return nil, 0, response.WrapError(response.CodeDatabaseError, "Failed to list comments", err)
	}

	return responsedto.ToCommentListResponse(items), total, nil
}

func (s *commentService) Update(ctx context.Context, input servicedto.UpdateCommentInput) (*responsedto.CommentResponse, error) {
	input.Content = strings.TrimSpace(input.Content)

	if err := validator.ValidateStruct(input); err != nil {
		return nil, err
	}

	comment, err := s.commentRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, response.WrapError(response.CodeDatabaseError, "Failed to fetch comment", err)
	}
	if comment == nil {
		return nil, response.NewError(response.CodeNotFound, "Comment not found")
	}

	// Authorization: hanya pembuat komentar yang boleh mengedit komentarnya
	if comment.AuthorID != input.ActorID {
		return nil, response.NewError(response.CodeForbidden, "You are not allowed to update this comment")
	}

	comment.Content = input.Content
	if err := s.commentRepo.Update(ctx, comment); err != nil {
		return nil, response.WrapError(response.CodeDatabaseError, "Failed to update comment", err)
	}

	logger.Info(ctx, "comment updated", "comment_id", comment.ID, "actor_id", input.ActorID)
	return responsedto.ToCommentResponse(comment), nil
}

func (s *commentService) Delete(ctx context.Context, input servicedto.DeleteCommentInput) error {
	if err := validator.ValidateStruct(input); err != nil {
		return err
	}

	comment, err := s.commentRepo.FindByID(ctx, input.ID)
	if err != nil {
		return response.WrapError(response.CodeDatabaseError, "Failed to fetch comment", err)
	}
	if comment == nil {
		return response.NewError(response.CodeNotFound, "Comment not found")
	}

	// Authorization: pembuat komentar ATAU role 'admin' boleh menghapus
	if comment.AuthorID != input.ActorID && !strings.EqualFold(input.ActorRole, model.RoleAdmin) {
		return response.NewError(response.CodeForbidden, "You are not allowed to delete this comment")
	}

	// Database Transaction: hapus komentar sekaligus kurangi counter comments_count pada post secara atomik
	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		deleted, err := s.commentRepo.Delete(txCtx, comment.ID)
		if err != nil {
			return response.WrapError(response.CodeDatabaseError, "Failed to delete comment", err)
		}
		if !deleted {
			return response.NewError(response.CodeNotFound, "Comment not found")
		}

		if err := s.postRepo.AdjustCommentsCount(txCtx, comment.PostID, -1); err != nil {
			return response.WrapError(response.CodeDatabaseError, "Failed to decrement post comments count", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	logger.Info(ctx, "comment deleted", "comment_id", comment.ID, "post_id", comment.PostID, "actor_id", input.ActorID)
	return nil
}

