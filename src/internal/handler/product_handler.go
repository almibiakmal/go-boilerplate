package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	requestdto "maj-core-srvc/src/internal/dto/request"
	_ "maj-core-srvc/src/internal/dto/response"
	"maj-core-srvc/src/internal/pkg/response"
	"maj-core-srvc/src/internal/pkg/validator"
	"maj-core-srvc/src/internal/service"
)

type ProductHandler struct {
	svc service.ProductService
}

func NewProductHandler(svc service.ProductService) *ProductHandler {
	return &ProductHandler{svc: svc}
}

// Create membuat produk baru.
//
// @Summary      Create Product
// @Description  Membuat data produk baru
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        request  body      requestdto.CreateProductRequest  true  "Create Product Payload"
// @Success      201      {object}  response.Envelope{data=responsedto.ProductResponse}
// @Failure      400      {object}  response.Envelope
// @Failure      409      {object}  response.Envelope
// @Failure      422      {object}  response.Envelope
// @Router       /api/v1/products [post]
func (h *ProductHandler) Create(c *fiber.Ctx) error {
	var req requestdto.CreateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return response.NewError(response.CodeBadRequest, "Invalid JSON request body")
	}

	if err := validator.ValidateStruct(req); err != nil {
		return err
	}

	res, err := h.svc.Create(c.UserContext(), req.ToServiceInput())
	if err != nil {
		return err
	}

	return response.Created(c, "Product created successfully", res)
}

// GetByID mengambil detail produk berdasarkan ID.
//
// @Summary      Get Product by ID
// @Description  Mengambil detail satu produk berdasarkan ID
// @Tags         Products
// @Produce      json
// @Param        id   path      int  true  "Product ID"
// @Success      200  {object}  response.Envelope{data=responsedto.ProductResponse}
// @Failure      400  {object}  response.Envelope
// @Failure      404  {object}  response.Envelope
// @Router       /api/v1/products/{id} [get]
func (h *ProductHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return response.NewError(response.CodeBadRequest, "Invalid product ID parameter")
	}

	res, err := h.svc.GetByID(c.UserContext(), id)
	if err != nil {
		return err
	}

	return response.OK(c, "Product retrieved successfully", res)
}

// List mengambil daftar produk dengan pagination & pencarian.
//
// @Summary      List Products
// @Description  Mengambil daftar produk dengan filter pencarian dan pagination
// @Tags         Products
// @Produce      json
// @Param        page      query     int     false  "Page number (default: 1)"
// @Param        per_page  query     int     false  "Items per page (default: 10, max: 100)"
// @Param        search    query     string  false  "Search by name or SKU"
// @Success      200       {object}  response.Envelope{data=[]responsedto.ProductResponse}
// @Failure      422       {object}  response.Envelope
// @Router       /api/v1/products [get]
func (h *ProductHandler) List(c *fiber.Ctx) error {
	req := requestdto.ListProductRequest{
		Page:    c.QueryInt("page", 1),
		PerPage: c.QueryInt("per_page", 10),
		Search:  c.Query("search", ""),
	}

	if err := validator.ValidateStruct(req); err != nil {
		return err
	}

	filter := req.ToServiceFilter()
	items, total, err := h.svc.List(c.UserContext(), filter)
	if err != nil {
		return err
	}

	return response.Paginated(c, "Products retrieved successfully", items, filter.Page, filter.PerPage, total)
}

// Update memperbarui data produk berdasarkan ID.
//
// @Summary      Update Product
// @Description  Memperbarui nama dan harga produk berdasarkan ID
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param        id       path      int                              true  "Product ID"
// @Param        request  body      requestdto.UpdateProductRequest  true  "Update Product Payload"
// @Success      200      {object}  response.Envelope{data=responsedto.ProductResponse}
// @Failure      400      {object}  response.Envelope
// @Failure      404      {object}  response.Envelope
// @Failure      422      {object}  response.Envelope
// @Router       /api/v1/products/{id} [put]
func (h *ProductHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return response.NewError(response.CodeBadRequest, "Invalid product ID parameter")
	}

	var req requestdto.UpdateProductRequest
	if err := c.BodyParser(&req); err != nil {
		return response.NewError(response.CodeBadRequest, "Invalid JSON request body")
	}

	if err := validator.ValidateStruct(req); err != nil {
		return err
	}

	res, err := h.svc.Update(c.UserContext(), req.ToServiceInput(id))
	if err != nil {
		return err
	}

	return response.OK(c, "Product updated successfully", res)
}

// Delete menghapus produk berdasarkan ID.
//
// @Summary      Delete Product
// @Description  Menghapus satu produk berdasarkan ID
// @Tags         Products
// @Produce      json
// @Param        id   path      int  true  "Product ID"
// @Success      200  {object}  response.Envelope
// @Failure      400  {object}  response.Envelope
// @Failure      404  {object}  response.Envelope
// @Router       /api/v1/products/{id} [delete]
func (h *ProductHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return response.NewError(response.CodeBadRequest, "Invalid product ID parameter")
	}

	if err := h.svc.Delete(c.UserContext(), id); err != nil {
		return err
	}

	return response.OK(c, "Product deleted successfully", nil)
}
