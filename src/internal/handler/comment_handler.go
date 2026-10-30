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

type CommentHandler struct {
	commentSvc service.CommentService
}

func NewCommentHandler(commentSvc service.CommentService) *CommentHandler {
	return &CommentHandler{commentSvc: commentSvc}
}

// Create menambahkan komentar baru pada sebuah artikel blog.
//
// @Summary      Create Comment on Post
// @Description  Menambahkan komentar baru pada post dan menambah counter comments_count dalam satu transaksi DB
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        postId   path      int                              true  "Post ID"
// @Param        request  body      requestdto.CreateCommentRequest  true  "Create Comment Payload"
// @Success      201      {object}  response.Envelope{data=responsedto.CommentResponse}
// @Failure      400      {object}  response.Envelope
// @Failure      401      {object}  response.Envelope
// @Failure      404      {object}  response.Envelope
// @Failure      422      {object}  response.Envelope
// @Router       /api/v1/posts/{postId}/comments [post]
func (h *CommentHandler) Create(c *fiber.Ctx) error {
	authorID, _, err := extractActor(c)
	if err != nil {
		return err
	}

	postID, err := strconv.ParseUint(c.Params("postId"), 10, 64)
	if err != nil || postID == 0 {
		return response.NewError(response.CodeBadRequest, "Invalid post ID parameter")
	}

	var req requestdto.CreateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.NewError(response.CodeBadRequest, "Invalid JSON request body")
	}

	if err := validator.ValidateStruct(req); err != nil {
		return err
	}

	res, err := h.commentSvc.Create(c.UserContext(), req.ToServiceInput(postID, authorID))
	if err != nil {
		return err
	}

	return response.Created(c, "Comment created successfully", res)
}

// ListByPostID mengambil daftar komentar pada sebuah artikel blog.
//
// @Summary      List Comments by Post ID
// @Description  Mengambil daftar komentar pada suatu artikel blog dengan pagination
// @Tags         Comments
// @Produce      json
// @Param        postId    path      int  true   "Post ID"
// @Param        page      query     int  false  "Page number (default: 1)"
// @Param        per_page  query     int  false  "Items per page (default: 10, max: 100)"
// @Success      200       {object}  response.Envelope{data=[]responsedto.CommentResponse}
// @Failure      400       {object}  response.Envelope
// @Failure      404       {object}  response.Envelope
// @Failure      422       {object}  response.Envelope
// @Router       /api/v1/posts/{postId}/comments [get]
func (h *CommentHandler) ListByPostID(c *fiber.Ctx) error {
	postID, err := strconv.ParseUint(c.Params("postId"), 10, 64)
	if err != nil || postID == 0 {
		return response.NewError(response.CodeBadRequest, "Invalid post ID parameter")
	}

	req := requestdto.ListCommentRequest{
		Page:    c.QueryInt("page", 1),
		PerPage: c.QueryInt("per_page", 10),
	}

	if err := validator.ValidateStruct(req); err != nil {
		return err
	}

	filter := req.ToServiceFilter(postID)
	items, total, err := h.commentSvc.ListByPostID(c.UserContext(), filter)
	if err != nil {
		return err
	}

	return response.Paginated(c, "Comments retrieved successfully", items, filter.Page, filter.PerPage, total)
}

// Update memperbarui isi komentar berdasarkan ID.
//
// @Summary      Update Comment
// @Description  Memperbarui isi komentar (hanya pembuat komentar yang diizinkan)
// @Tags         Comments
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id       path      int                              true  "Comment ID"
// @Param        request  body      requestdto.UpdateCommentRequest  true  "Update Comment Payload"
// @Success      200      {object}  response.Envelope{data=responsedto.CommentResponse}
// @Failure      400      {object}  response.Envelope
// @Failure      401      {object}  response.Envelope
// @Failure      403      {object}  response.Envelope
// @Failure      404      {object}  response.Envelope
// @Failure      422      {object}  response.Envelope
// @Router       /api/v1/comments/{id} [put]
func (h *CommentHandler) Update(c *fiber.Ctx) error {
	actorID, actorRole, err := extractActor(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return response.NewError(response.CodeBadRequest, "Invalid comment ID parameter")
	}

	var req requestdto.UpdateCommentRequest
	if err := c.BodyParser(&req); err != nil {
		return response.NewError(response.CodeBadRequest, "Invalid JSON request body")
	}

	if err := validator.ValidateStruct(req); err != nil {
		return err
	}

	res, err := h.commentSvc.Update(c.UserContext(), req.ToServiceInput(id, actorID, actorRole))
	if err != nil {
		return err
	}

	return response.OK(c, "Comment updated successfully", res)
}

// Delete menghapus komentar dan mengurangi counter comments_count pada post dalam satu transaksi DB.
//
// @Summary      Delete Comment
// @Description  Menghapus komentar (diizinkan untuk pembuat komentar atau role admin)
// @Tags         Comments
// @Produce      json
// @Security     BearerAuth
// @Param        id   path      int  true  "Comment ID"
// @Success      200  {object}  response.Envelope
// @Failure      400  {object}  response.Envelope
// @Failure      401  {object}  response.Envelope
// @Failure      403  {object}  response.Envelope
// @Failure      404  {object}  response.Envelope
// @Router       /api/v1/comments/{id} [delete]
func (h *CommentHandler) Delete(c *fiber.Ctx) error {
	actorID, actorRole, err := extractActor(c)
	if err != nil {
		return err
	}

	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return response.NewError(response.CodeBadRequest, "Invalid comment ID parameter")
	}

	err = h.commentSvc.Delete(c.UserContext(), servicedto.DeleteCommentInput{
		ID:        id,
		ActorID:   actorID,
		ActorRole: actorRole,
	})
	if err != nil {
		return err
	}

	return response.OK(c, "Comment deleted successfully", nil)
}

