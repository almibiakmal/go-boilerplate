package servicedto

// --- Auth & User Service DTOs ---

type RegisterUserInput struct {
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Email    string `json:"email" validate:"required,email,max=150"`
	Password string `json:"password" validate:"required,min=6,max=72"`
	Role     string `json:"role" validate:"omitempty,oneof=user admin"`
}

type LoginInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// --- Post Service DTOs ---

type CreatePostInput struct {
	AuthorID    uint64 `json:"author_id" validate:"required,gt=0"`
	Title       string `json:"title" validate:"required,min=3,max=200"`
	Content     string `json:"content" validate:"required,min=10"`
	IsPublished bool   `json:"is_published"`
}

type UpdatePostInput struct {
	ID          uint64 `json:"id" validate:"required,gt=0"`
	ActorID     uint64 `json:"actor_id" validate:"required,gt=0"`
	ActorRole   string `json:"actor_role"`
	Title       string `json:"title" validate:"required,min=3,max=200"`
	Content     string `json:"content" validate:"required,min=10"`
	IsPublished bool   `json:"is_published"`
}

type DeletePostInput struct {
	ID        uint64 `json:"id" validate:"required,gt=0"`
	ActorID   uint64 `json:"actor_id" validate:"required,gt=0"`
	ActorRole string `json:"actor_role"`
}

type ListPostFilter struct {
	Page     int
	PerPage  int
	Search   string
	AuthorID uint64
}

// --- Comment Service DTOs ---

type CreateCommentInput struct {
	PostID   uint64 `json:"post_id" validate:"required,gt=0"`
	AuthorID uint64 `json:"author_id" validate:"required,gt=0"`
	Content  string `json:"content" validate:"required,min=2,max=2000"`
}

type UpdateCommentInput struct {
	ID        uint64 `json:"id" validate:"required,gt=0"`
	ActorID   uint64 `json:"actor_id" validate:"required,gt=0"`
	ActorRole string `json:"actor_role"`
	Content   string `json:"content" validate:"required,min=2,max=2000"`
}

type DeleteCommentInput struct {
	ID        uint64 `json:"id" validate:"required,gt=0"`
	ActorID   uint64 `json:"actor_id" validate:"required,gt=0"`
	ActorRole string `json:"actor_role"`
}

type ListCommentFilter struct {
	PostID  uint64
	Page    int
	PerPage int
}

