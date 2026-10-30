package requestdto

import (
	"strings"

	servicedto "maj-core-srvc/src/internal/dto/service"
)

// --- Auth Requests ---

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email,max=150"`
	Password string `json:"password" validate:"required,min=6,max=72"`
}

func (r RegisterRequest) ToServiceInput() servicedto.RegisterUserInput {
	return servicedto.RegisterUserInput{
		Name:     strings.TrimSpace(r.Name),
		Email:    strings.ToLower(strings.TrimSpace(r.Email)),
		Password: r.Password,
	}
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (r LoginRequest) ToServiceInput() servicedto.LoginInput {
	return servicedto.LoginInput{
		Email:    strings.ToLower(strings.TrimSpace(r.Email)),
		Password: r.Password,
	}
}

// --- Post Requests ---

type CreatePostRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=200"`
	Content     string `json:"content" validate:"required,min=10"`
	IsPublished *bool  `json:"is_published"`
}

func (r CreatePostRequest) ToServiceInput(authorID uint64) servicedto.CreatePostInput {
	isPublished := true
	if r.IsPublished != nil {
		isPublished = *r.IsPublished
	}
	return servicedto.CreatePostInput{
		AuthorID:    authorID,
		Title:       strings.TrimSpace(r.Title),
		Content:     strings.TrimSpace(r.Content),
		IsPublished: isPublished,
	}
}

type UpdatePostRequest struct {
	Title       string `json:"title" validate:"required,min=3,max=200"`
	Content     string `json:"content" validate:"required,min=10"`
	IsPublished *bool  `json:"is_published"`
}

func (r UpdatePostRequest) ToServiceInput(id, actorID uint64, actorRole string) servicedto.UpdatePostInput {
	isPublished := true
	if r.IsPublished != nil {
		isPublished = *r.IsPublished
	}
	return servicedto.UpdatePostInput{
		ID:          id,
		ActorID:     actorID,
		ActorRole:   actorRole,
		Title:       strings.TrimSpace(r.Title),
		Content:     strings.TrimSpace(r.Content),
		IsPublished: isPublished,
	}
}

type ListPostRequest struct {
	Page     int    `query:"page" validate:"omitempty,gte=1"`
	PerPage  int    `query:"per_page" validate:"omitempty,gte=1,lte=100"`
	Search   string `query:"search" validate:"omitempty,max=100"`
	AuthorID uint64 `query:"author_id" validate:"omitempty,gte=1"`
}

func (r ListPostRequest) ToServiceFilter() servicedto.ListPostFilter {
	return servicedto.ListPostFilter{
		Page:     r.Page,
		PerPage:  r.PerPage,
		Search:   strings.TrimSpace(r.Search),
		AuthorID: r.AuthorID,
	}
}

// --- Comment Requests ---

type CreateCommentRequest struct {
	Content string `json:"content" validate:"required,min=2,max=2000"`
}

func (r CreateCommentRequest) ToServiceInput(postID, authorID uint64) servicedto.CreateCommentInput {
	return servicedto.CreateCommentInput{
		PostID:   postID,
		AuthorID: authorID,
		Content:  strings.TrimSpace(r.Content),
	}
}

type UpdateCommentRequest struct {
	Content string `json:"content" validate:"required,min=2,max=2000"`
}

func (r UpdateCommentRequest) ToServiceInput(id, actorID uint64, actorRole string) servicedto.UpdateCommentInput {
	return servicedto.UpdateCommentInput{
		ID:        id,
		ActorID:   actorID,
		ActorRole: actorRole,
		Content:   strings.TrimSpace(r.Content),
	}
}

type ListCommentRequest struct {
	Page    int `query:"page" validate:"omitempty,gte=1"`
	PerPage int `query:"per_page" validate:"omitempty,gte=1,lte=100"`
}

func (r ListCommentRequest) ToServiceFilter(postID uint64) servicedto.ListCommentFilter {
	return servicedto.ListCommentFilter{
		PostID:  postID,
		Page:    r.Page,
		PerPage: r.PerPage,
	}
}

