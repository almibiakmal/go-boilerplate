package handler

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	requestdto "maj-core-srvc/src/internal/dto/request"
	_ "maj-core-srvc/src/internal/dto/response"
	"maj-core-srvc/src/internal/middleware"
	"maj-core-srvc/src/internal/pkg/response"
	"maj-core-srvc/src/internal/pkg/validator"
	"maj-core-srvc/src/internal/service"
)

type AuthHandler struct {
	authSvc service.AuthService
}

func NewAuthHandler(authSvc service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// Register mendaftarkan user baru dan mengembalikan JWT token.
//
// @Summary      Register User
// @Description  Mendaftarkan akun user baru dan mengembalikan JWT access token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      requestdto.RegisterRequest  true  "Register Payload"
// @Success      201      {object}  response.Envelope{data=responsedto.AuthResponse}
// @Failure      400      {object}  response.Envelope
// @Failure      409      {object}  response.Envelope
// @Failure      422      {object}  response.Envelope
// @Router       /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *fiber.Ctx) error {
	var req requestdto.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		return response.NewError(response.CodeBadRequest, "Invalid JSON request body")
	}

	if err := validator.ValidateStruct(req); err != nil {
		return err
	}

	res, err := h.authSvc.Register(c.UserContext(), req.ToServiceInput())
	if err != nil {
		return err
	}

	return response.Created(c, "User registered successfully", res)
}

// Login mengautentikasi user dengan email & password lalu mengembalikan JWT token.
//
// @Summary      Login User
// @Description  Autentikasi user dengan email dan password untuk mendapatkan JWT access token
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param        request  body      requestdto.LoginRequest  true  "Login Payload"
// @Success      200      {object}  response.Envelope{data=responsedto.AuthResponse}
// @Failure      400      {object}  response.Envelope
// @Failure      401      {object}  response.Envelope
// @Failure      422      {object}  response.Envelope
// @Router       /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req requestdto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return response.NewError(response.CodeBadRequest, "Invalid JSON request body")
	}

	if err := validator.ValidateStruct(req); err != nil {
		return err
	}

	res, err := h.authSvc.Login(c.UserContext(), req.ToServiceInput())
	if err != nil {
		return err
	}

	return response.OK(c, "Login successful", res)
}

// Me mengambil profil user yang sedang terautentikasi.
//
// @Summary      Get Current User Profile
// @Description  Mengambil data profil user yang sedang login berdasarkan Bearer JWT token
// @Tags         Auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  response.Envelope{data=responsedto.UserResponse}
// @Failure      401  {object}  response.Envelope
// @Failure      404  {object}  response.Envelope
// @Router       /api/v1/auth/me [get]
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	userID, _, err := extractActor(c)
	if err != nil {
		return err
	}

	res, err := h.authSvc.GetProfile(c.UserContext(), userID)
	if err != nil {
		return err
	}

	return response.OK(c, "User profile retrieved successfully", res)
}

// extractActor mengambil user_id (uint64) dan role dari JWT claims pada Fiber context.
func extractActor(c *fiber.Ctx) (uint64, string, error) {
	claims, ok := middleware.GetAuthClaims(c)
	if !ok || claims == nil {
		return 0, "", response.NewError(response.CodeUnauthorized, "Unauthorized access")
	}

	userID, err := strconv.ParseUint(claims.UserID, 10, 64)
	if err != nil || userID == 0 {
		return 0, "", response.NewError(response.CodeUnauthorized, "Invalid user identity in token")
	}

	return userID, claims.Role, nil
}

