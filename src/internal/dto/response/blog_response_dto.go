package responsedto

import (
	"time"

	"maj-core-srvc/src/internal/model"
)

// UserResponse merepresentasikan data profil user pada JSON response (tanpa password_hash).
type UserResponse struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

func ToUserResponse(u *model.User) *UserResponse {
	if u == nil {
		return nil
	}
	return &UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}

// AuthResponse merepresentasikan response login/registrasi beserta JWT access token.
type AuthResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int64        `json:"expires_in"`
	User        UserResponse `json:"user"`
}

// AuthorSummary adalah ringkasan author yang disematkan pada Post/Comment response.
type AuthorSummary struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

// PostResponse merepresentasikan data artikel blog pada JSON response.
type PostResponse struct {
	ID            uint64         `json:"id"`
	AuthorID      uint64         `json:"author_id"`
	Author        *AuthorSummary `json:"author,omitempty"`
	Title         string         `json:"title"`
	Slug          string         `json:"slug"`
	Content       string         `json:"content"`
	IsPublished   bool           `json:"is_published"`
	CommentsCount int            `json:"comments_count"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func ToPostResponse(p *model.Post) *PostResponse {
	if p == nil {
		return nil
	}
	res := &PostResponse{
		ID:            p.ID,
		AuthorID:      p.AuthorID,
		Title:         p.Title,
		Slug:          p.Slug,
		Content:       p.Content,
		IsPublished:   p.IsPublished,
		CommentsCount: p.CommentsCount,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
	if p.Author != nil && p.Author.ID != 0 {
		res.Author = &AuthorSummary{
			ID:   p.Author.ID,
			Name: p.Author.Name,
		}
	}
	return res
}

func ToPostListResponse(items []model.Post) []PostResponse {
	res := make([]PostResponse, len(items))
	for i := range items {
		res[i] = *ToPostResponse(&items[i])
	}
	return res
}

// CommentResponse merepresentasikan data komentar pada JSON response.
type CommentResponse struct {
	ID        uint64         `json:"id"`
	PostID    uint64         `json:"post_id"`
	AuthorID  uint64         `json:"author_id"`
	Author    *AuthorSummary `json:"author,omitempty"`
	Content   string         `json:"content"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func ToCommentResponse(c *model.Comment) *CommentResponse {
	if c == nil {
		return nil
	}
	res := &CommentResponse{
		ID:        c.ID,
		PostID:    c.PostID,
		AuthorID:  c.AuthorID,
		Content:   c.Content,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
	if c.Author != nil && c.Author.ID != 0 {
		res.Author = &AuthorSummary{
			ID:   c.Author.ID,
			Name: c.Author.Name,
		}
	}
	return res
}

func ToCommentListResponse(items []model.Comment) []CommentResponse {
	res := make([]CommentResponse, len(items))
	for i := range items {
		res[i] = *ToCommentResponse(&items[i])
	}
	return res
}

