package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"maj-core-srvc/src/internal/config"
	responsedto "maj-core-srvc/src/internal/dto/response"
	servicedto "maj-core-srvc/src/internal/dto/service"
	"maj-core-srvc/src/internal/infra/logger"
	"maj-core-srvc/src/internal/middleware"
	"maj-core-srvc/src/internal/model"
	"maj-core-srvc/src/internal/pkg/response"
	"maj-core-srvc/src/internal/pkg/validator"
	"maj-core-srvc/src/internal/repository"
)

type AuthService interface {
	Register(ctx context.Context, input servicedto.RegisterUserInput) (*responsedto.AuthResponse, error)
	Login(ctx context.Context, input servicedto.LoginInput) (*responsedto.AuthResponse, error)
	GetProfile(ctx context.Context, userID uint64) (*responsedto.UserResponse, error)
}

type authService struct {
	userRepo repository.UserRepository
	authCfg  config.AuthConfig
}

func NewAuthService(userRepo repository.UserRepository, authCfg config.AuthConfig) AuthService {
	return &authService{
		userRepo: userRepo,
		authCfg:  authCfg,
	}
}

func (s *authService) Register(ctx context.Context, input servicedto.RegisterUserInput) (*responsedto.AuthResponse, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))
	if input.Role == "" {
		input.Role = model.RoleUser
	}

	if err := validator.ValidateStruct(input); err != nil {
		return nil, err
	}

	existing, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, response.WrapError(response.CodeDatabaseError, "Failed to check existing email", err)
	}
	if existing != nil {
		return nil, response.NewError(response.CodeDuplicateEntry, "Email is already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, response.WrapError(response.CodeInternalError, "Failed to hash password", err)
	}

	user := &model.User{
		Name:     input.Name,
		Email:    input.Email,
		Password: string(hash),
		Role:     input.Role,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, response.WrapError(response.CodeDatabaseError, "Failed to register user", err)
	}

	logger.Info(ctx, "user registered", "user_id", user.ID, "email", user.Email, "role", user.Role)
	return s.buildAuthResponse(user)
}

func (s *authService) Login(ctx context.Context, input servicedto.LoginInput) (*responsedto.AuthResponse, error) {
	input.Email = strings.ToLower(strings.TrimSpace(input.Email))

	if err := validator.ValidateStruct(input); err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByEmail(ctx, input.Email)
	if err != nil {
		return nil, response.WrapError(response.CodeDatabaseError, "Failed to find user", err)
	}
	if user == nil {
		return nil, response.NewError(response.CodeUnauthorized, "Invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return nil, response.NewError(response.CodeUnauthorized, "Invalid email or password")
	}

	logger.Info(ctx, "user logged in", "user_id", user.ID, "email", user.Email)
	return s.buildAuthResponse(user)
}

func (s *authService) GetProfile(ctx context.Context, userID uint64) (*responsedto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, response.WrapError(response.CodeDatabaseError, "Failed to fetch user profile", err)
	}
	if user == nil {
		return nil, response.NewError(response.CodeNotFound, "User not found")
	}
	return responsedto.ToUserResponse(user), nil
}

func (s *authService) buildAuthResponse(user *model.User) (*responsedto.AuthResponse, error) {
	expDuration := s.authCfg.JWTExpiration
	if expDuration <= 0 {
		expDuration = 24 * time.Hour
	}

	now := time.Now()
	claims := middleware.AuthClaims{
		UserID: strconv.FormatUint(user.ID, 10),
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.authCfg.JWTIssuer,
			Subject:   strconv.FormatUint(user.ID, 10),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expDuration)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(s.authCfg.JWTSecret))
	if err != nil {
		return nil, response.WrapError(response.CodeInternalError, "Failed to sign access token", err)
	}

	return &responsedto.AuthResponse{
		AccessToken: signed,
		TokenType:   "Bearer",
		ExpiresIn:   int64(expDuration.Seconds()),
		User:        *responsedto.ToUserResponse(user),
	}, nil
}
