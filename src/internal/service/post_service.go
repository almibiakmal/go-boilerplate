package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	responsedto "maj-core-srvc/src/internal/dto/response"
	servicedto "maj-core-srvc/src/internal/dto/service"
	"maj-core-srvc/src/internal/infra/database"
	"maj-core-srvc/src/internal/infra/logger"
	"maj-core-srvc/src/internal/model"
	"maj-core-srvc/src/internal/pkg/response"
	"maj-core-srvc/src/internal/pkg/validator"
	"maj-core-srvc/src/internal/repository"
)

var nonSlugChar = regexp.MustCompile(`[^a-z0-9]+`)

type PostService interface {
	Create(ctx context.Context, input servicedto.CreatePostInput) (*responsedto.PostResponse, error)
	GetByID(ctx context.Context, id uint64) (*responsedto.PostResponse, error)
	List(ctx context.Context, filter servicedto.ListPostFilter) ([]responsedto.PostResponse, int64, error)
	Update(ctx context.Context, input servicedto.UpdatePostInput) (*responsedto.PostResponse, error)
	Delete(ctx context.Context, input servicedto.DeletePostInput) error
}

type postService struct {
	txManager   database.TxManager
	postRepo    repository.PostRepository
	commentRepo repository.CommentRepository
}

func NewPostService(
	txManager database.TxManager,
	postRepo repository.PostRepository,
	commentRepo repository.CommentRepository,
) PostService {
	return &postService{
		txManager:   txManager,
		postRepo:    postRepo,
		commentRepo: commentRepo,
	}
}

func (s *postService) Create(ctx context.Context, input servicedto.CreatePostInput) (*responsedto.PostResponse, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Content = strings.TrimSpace(input.Content)

	if err := validator.ValidateStruct(input); err != nil {
		return nil, err
	}

	post := &model.Post{
		AuthorID:    input.AuthorID,
		Title:       input.Title,
		Slug:        generateUniqueSlug(input.Title),
		Content:     input.Content,
		IsPublished: input.IsPublished,
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, response.WrapError(response.CodeDatabaseError, "Failed to create post", err)
	}

	created, err := s.postRepo.FindByID(ctx, post.ID)
	if err != nil || created == nil {
		logger.Info(ctx, "post created", "post_id", post.ID, "author_id", post.AuthorID)
		return responsedto.ToPostResponse(post), nil
	}

	logger.Info(ctx, "post created", "post_id", created.ID, "author_id", created.AuthorID)
	return responsedto.ToPostResponse(created), nil
}

func (s *postService) GetByID(ctx context.Context, id uint64) (*responsedto.PostResponse, error) {
	post, err := s.postRepo.FindByID(ctx, id)
	if err != nil {
		return nil, response.WrapError(response.CodeDatabaseError, "Failed to fetch post", err)
	}
	if post == nil {
		return nil, response.NewError(response.CodeNotFound, "Post not found")
	}
	return responsedto.ToPostResponse(post), nil
}

func (s *postService) List(ctx context.Context, filter servicedto.ListPostFilter) ([]responsedto.PostResponse, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 10
	}

	items, total, err := s.postRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, response.WrapError(response.CodeDatabaseError, "Failed to list posts", err)
	}
	return responsedto.ToPostListResponse(items), total, nil
}

func (s *postService) Update(ctx context.Context, input servicedto.UpdatePostInput) (*responsedto.PostResponse, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Content = strings.TrimSpace(input.Content)

	if err := validator.ValidateStruct(input); err != nil {
		return nil, err
	}

	post, err := s.postRepo.FindByID(ctx, input.ID)
	if err != nil {
		return nil, response.WrapError(response.CodeDatabaseError, "Failed to fetch post", err)
	}
	if post == nil {
		return nil, response.NewError(response.CodeNotFound, "Post not found")
	}

	// Authorization: hanya pemilik post yang boleh mengedit
	if post.AuthorID != input.ActorID {
		return nil, response.NewError(response.CodeForbidden, "You are not allowed to update this post")
	}

	post.Title = input.Title
	post.Content = input.Content
	post.IsPublished = input.IsPublished

	if err := s.postRepo.Update(ctx, post); err != nil {
		return nil, response.WrapError(response.CodeDatabaseError, "Failed to update post", err)
	}

	logger.Info(ctx, "post updated", "post_id", post.ID, "actor_id", input.ActorID)
	return responsedto.ToPostResponse(post), nil
}

func (s *postService) Delete(ctx context.Context, input servicedto.DeletePostInput) error {
	if err := validator.ValidateStruct(input); err != nil {
		return err
	}

	post, err := s.postRepo.FindByID(ctx, input.ID)
	if err != nil {
		return response.WrapError(response.CodeDatabaseError, "Failed to fetch post", err)
	}
	if post == nil {
		return response.NewError(response.CodeNotFound, "Post not found")
	}

	// Anti-IDOR Authorization: hanya pemilik post (AuthorID == ActorID) yang boleh menghapus post miliknya
	if post.AuthorID != input.ActorID {
		return response.NewError(response.CodeForbidden, "You are not allowed to delete this post")
	}

	// Database Transaction: hapus semua komentar milik post sekaligus hapus post secara atomik (dikunci juga dengan author_id di level SQL)
	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		if err := s.commentRepo.DeleteByPostID(txCtx, post.ID); err != nil {
			return response.WrapError(response.CodeDatabaseError, "Failed to delete post comments", err)
		}
		deleted, err := s.postRepo.Delete(txCtx, post.ID, input.ActorID)
		if err != nil {
			return response.WrapError(response.CodeDatabaseError, "Failed to delete post", err)
		}
		if !deleted {
			return response.NewError(response.CodeNotFound, "Post not found")
		}
		return nil
	})
	if err != nil {
		return err
	}

	logger.Info(ctx, "post deleted", "post_id", post.ID, "actor_id", input.ActorID)
	return nil
}

func generateUniqueSlug(title string) string {
	base := strings.Trim(nonSlugChar.ReplaceAllString(strings.ToLower(title), "-"), "-")
	if base == "" {
		base = "post"
	}
	return fmt.Sprintf("%s-%d", base, time.Now().UnixNano())
}

