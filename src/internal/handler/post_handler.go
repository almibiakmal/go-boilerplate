package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	requestdto "maj-core-srvc/src/internal/dto/request"
	_ "maj-core-srvc/src/internal/dto/response"
	servicedto "maj-core-srvc/src/internal/dto/service"
	"maj-core-srvc/src/internal/pkg/response"
	"maj-core-srvc/src/internal/pkg/validator"
	"maj-core-srvc/src/internal/service"
)

type PostHandler struct {
	postSvc service.PostService
}

func NewPostHandler(postSvc service.PostService) *PostHandler {
	return &PostHandler{postSvc: postSvc}
}

// Create membuat artikel blog baru.
//
// @Summary      Create Post
// @Description  Membuat artikel blog baru (wajib login Bearer token)
// @Tags         Posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request  body      requestdto.CreatePostRequest  true  "Create Post Payload"
// @Success      201      {object}  response.Envelope{data=responsedto.PostResponse}
// @Failure      400      {object}  response.Envelope
// @Failure      401      {object}  response.Envelope
// @Failure      422      {object}  response.Envelope
// @Router       /api/v1/posts [post]
func (h *PostHandler) Create(c *fiber.Ctx) error {
	authorID, _, err := extractActor(c)
	if err != nil {
		return err
	}

	var req requestdto.CreatePostRequest
	if err := c.BodyParser(&req); err != nil {
		return response.NewError(response.CodeBadRequest, "Invalid JSON request body")
	}

	if err := validator.ValidateStruct(req); err != nil {
		return err
	}

	res, err := h.postSvc.Create(c.UserContext(), req.ToServiceInput(authorID))
	if err != nil {
		return err
	}

	return response.Created(c, "Post created successfully", res)
}

// GetByID mengambil detail satu artikel blog berdasarkan ID.
//
// @Summary      Get Post by ID
// @Description  Mengambil detail satu artikel blog beserta informasi author
// @Tags         Posts
// @Produce      json
// @Param        id   path      int  true  "Post ID"
// @Success      200  {object}  response.Envelope{data=responsedto.PostResponse}
// @Failure      400  {object}  response.Envelope
// @Failure      404  {object}  response.Envelope
// @Router       /api/v1/posts/{id} [get]
func (h *PostHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return response.NewError(response.CodeBadRequest, "Invalid post ID parameter")
	}

	res, err := h.postSvc.GetByID(c.UserContext(), id)
	if err != nil {
		return err
	}

	return response.OK(c, "Post retrieved successfully", res)
}

// List mengambil daftar artikel blog dengan pagination & pencarian.
//
// @Summary      List Posts
// @Description  Mengambil daftar artikel blog dengan filter pencarian, author_id, dan pagination
// @Tags         Posts
// @Produce      json
// @Param        page       query     int     false  "Page number (default: 1)"
// @Param        per_page   query     int     false  "Items per page (default: 10, max: 100)"
// @Param        search     query     string  false  "Search by title or content"
// @Param        author_id  query     int     false  "Filter by Author ID"
// @Success      200        {object}  response.Envelope{data=[]responsedto.PostResponse}
// @Failure      422        {object}  response.Envelope
// @Router       /api/v1/posts [get]
func (h *PostHandler) List(c *fiber.Ctx) error {
	req := requestdto.ListPostRequest{
		Page:     c.QueryInt("page", 1),
		PerPage:  c.QueryInt("per_page", 10),
		Search:   c.Query("search", ""),
		AuthorID: uint64(c.QueryInt("author_id", 0)),
	}

	if err := validator.ValidateStruct(req); err != nil {
		return err
	}

	filter := req.ToServiceFilter()
	items, total, err := h.postSvc.List(c.UserContext(), filter)
	if err != nil {
		return err
	}

	return response.Paginated(c, "Posts retrieved successfully", items, filter.Page, filter.PerPage, total)
}

// Update memperbarui artikel blog berdasarkan ID.
//
// @Summary      Update Post
// @Description  Memperbarui artikel blog (hanya pemilik post yang diizinkan)
// @Tags         Posts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int                           true  "Post ID"
// @Param        request  body      requestdto.UpdatePostRequest  true  "Update Post Payload"
// @Success      200      {object}  response.Envelope{data=responsedto.PostResponse}
// @Failure      400      {object}  response.Envelope
// @Failure      401      {object}  response.Envelope
// @Failure      403      {object}  response.Envelope
// @Failure      404      {object}  response.Envelope
// @Failure      422      {object}  response.Envelope
// @Router       /api/v1/posts/{id} [put]
func (h *PostHandler) Update(c *fiber.Ctx) error {
	actorID, actorRole, err := extractActor(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return response.NewError(response.CodeBadRequest, "Invalid post ID parameter")
	}

	var req requestdto.UpdatePostRequest
	if err := c.BodyParser(&req); err != nil {
		return response.NewError(response.CodeBadRequest, "Invalid JSON request body")
	}

	if err := validator.ValidateStruct(req); err != nil {
		return err
	}

	res, err := h.postSvc.Update(c.UserContext(), req.ToServiceInput(id, actorID, actorRole))
	if err != nil {
		return err
	}

	return response.OK(c, "Post updated successfully", res)
}

// Delete menghapus artikel blog beserta seluruh komentarnya dalam satu transaksi database.
//
// @Summary      Delete Post
// @Description  Menghapus artikel blog beserta komentarnya (hanya pemilik post yang diizinkan)
// @Tags         Posts
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Post ID"
// @Success      200  {object}  response.Envelope
// @Failure      400  {object}  response.Envelope
// @Failure      401  {object}  response.Envelope
// @Failure      403  {object}  response.Envelope
// @Failure      404  {object}  response.Envelope
// @Router       /api/v1/posts/{id} [delete]
func (h *PostHandler) Delete(c *fiber.Ctx) error {
	actorID, actorRole, err := extractActor(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return response.NewError(response.CodeBadRequest, "Invalid post ID parameter")
	}

	err = h.postSvc.Delete(c.UserContext(), servicedto.DeletePostInput{
		ID:        id,
		ActorID:   actorID,
		ActorRole: actorRole,
	})
	if err != nil {
		return err
	}

	return response.OK(c, "Post deleted successfully", nil)
}

