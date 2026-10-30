package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"maj-core-srvc/src/internal/config"
	responsedto "maj-core-srvc/src/internal/dto/response"
	"maj-core-srvc/src/internal/handler"
	"maj-core-srvc/src/internal/infra/database"
	"maj-core-srvc/src/internal/middleware"
	"maj-core-srvc/src/internal/model"
	"maj-core-srvc/src/internal/pkg/response"
	"maj-core-srvc/src/internal/repository"
	"maj-core-srvc/src/internal/router"
	"maj-core-srvc/src/internal/service"
)

func setupBlogTestApp(t *testing.T) *fiber.App {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Post{}, &model.Comment{}))

	cfg := &config.Config{
		Auth: config.AuthConfig{
			JWTSecret:      "test-jwt-secret",
			JWTIssuer:      "maj-core-srvc",
			JWTExpiration:  time.Hour,
			InternalAPIKey: "test-internal-key",
		},
	}

	txManager := database.NewTxManager(db)
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)

	authSvc := service.NewAuthService(userRepo, cfg.Auth)
	postSvc := service.NewPostService(txManager, postRepo, commentRepo)
	commentSvc := service.NewCommentService(txManager, postRepo, commentRepo)

	app := fiber.New(fiber.Config{
		ErrorHandler: middleware.ErrorHandler,
	})

	router.RegisterHTTP(app, router.Dependencies{
		Config:         cfg,
		HealthHandler:  handler.NewHealthHandler(db),
		AuthHandler:    handler.NewAuthHandler(authSvc),
		PostHandler:    handler.NewPostHandler(postSvc),
		CommentHandler: handler.NewCommentHandler(commentSvc),
	})

	return app
}

func TestBlogHTTP_EndToEnd(t *testing.T) {
	app := setupBlogTestApp(t)

	// 1. Register User (201 Created)
	regBody := []byte(`{"name":"Andi Wijaya","email":"andi@majoo.id","password":"password123"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(regBody))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var regEnv struct {
		Success bool                     `json:"success"`
		Code    response.Code            `json:"code"`
		Data    responsedto.AuthResponse `json:"data"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&regEnv))
	assert.True(t, regEnv.Success)
	token := regEnv.Data.AccessToken
	require.NotEmpty(t, token)

	// 2. Create Post Tanpa Token -> 401 Unauthorized
	postBody := []byte(`{"title":"Belajar Golang Fiber","content":"Panduan lengkap membuat REST API dengan Fiber dan GORM."}`)
	unauthReq := httptest.NewRequest(http.MethodPost, "/api/v1/posts", bytes.NewReader(postBody))
	unauthReq.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

	unauthResp, err := app.Test(unauthReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, unauthResp.StatusCode)

	// 3. Create Post Dengan Token -> 201 Created
	createPostReq := httptest.NewRequest(http.MethodPost, "/api/v1/posts", bytes.NewReader(postBody))
	createPostReq.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	createPostReq.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)

	createPostResp, err := app.Test(createPostReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createPostResp.StatusCode)

	var postEnv struct {
		Success bool                     `json:"success"`
		Data    responsedto.PostResponse `json:"data"`
	}
	require.NoError(t, json.NewDecoder(createPostResp.Body).Decode(&postEnv))
	postID := postEnv.Data.ID
	assert.Equal(t, uint64(1), postID)

	// 4. Create Comment pada Post tersebut -> 201 Created
	commentBody := []byte(`{"content":"Mantap sekali pembahasannya!"}`)
	createCommentReq := httptest.NewRequest(
		http.MethodPost,
		fmt.Sprintf("/api/v1/posts/%d/comments", postID),
		bytes.NewReader(commentBody),
	)
	createCommentReq.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	createCommentReq.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)

	createCommentResp, err := app.Test(createCommentReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, createCommentResp.StatusCode)

	// 5. Get Post Detail -> comments_count harus sudah bernilai 1
	getPostReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/posts/%d", postID), nil)
	getPostResp, err := app.Test(getPostReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getPostResp.StatusCode)
	require.NoError(t, json.NewDecoder(getPostResp.Body).Decode(&postEnv))
	assert.Equal(t, 1, postEnv.Data.CommentsCount)

	// 6. Anti-IDOR Check via HTTP: User ke-2 mencoba menghapus Post & Comment milik User ke-1 -> harus ditolak 403 Forbidden
	reg2Body := []byte(`{"name":"Budi Hacker","email":"budi@majoo.id","password":"password123"}`)
	reg2Req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewReader(reg2Body))
	reg2Req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	reg2Resp, err := app.Test(reg2Req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, reg2Resp.StatusCode)

	var reg2Env struct {
		Data responsedto.AuthResponse `json:"data"`
	}
	require.NoError(t, json.NewDecoder(reg2Resp.Body).Decode(&reg2Env))
	attackerToken := reg2Env.Data.AccessToken

	idorDeletePostReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/posts/%d", postID), nil)
	idorDeletePostReq.Header.Set(fiber.HeaderAuthorization, "Bearer "+attackerToken)
	idorDeletePostResp, err := app.Test(idorDeletePostReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, idorDeletePostResp.StatusCode)

	idorDeleteCommentReq := httptest.NewRequest(http.MethodDelete, "/api/v1/comments/1", nil)
	idorDeleteCommentReq.Header.Set(fiber.HeaderAuthorization, "Bearer "+attackerToken)
	idorDeleteCommentResp, err := app.Test(idorDeleteCommentReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, idorDeleteCommentResp.StatusCode)

	// 7. Uji endpoint lainnya: Login, Me, List Posts, Update Post, List Comments, Get Comment, Update Comment, Delete Comment, Delete Post, Healthz, Readyz
	loginBody := []byte(`{"email":"andi@majoo.id","password":"password123"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewReader(loginBody))
	loginReq.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	loginResp, err := app.Test(loginReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)

	meReq := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	meReq.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)
	meResp, err := app.Test(meReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, meResp.StatusCode)

	listPostsReq := httptest.NewRequest(http.MethodGet, "/api/v1/posts?page=1&per_page=10&search=Golang", nil)
	listPostsResp, err := app.Test(listPostsReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, listPostsResp.StatusCode)

	updatePostBody := []byte(`{"title":"Belajar Golang Fiber Updated","content":"Panduan lengkap membuat REST API dengan Fiber dan GORM updated.","is_published":true}`)
	updatePostReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/v1/posts/%d", postID), bytes.NewReader(updatePostBody))
	updatePostReq.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	updatePostReq.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)
	updatePostResp, err := app.Test(updatePostReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, updatePostResp.StatusCode)

	listCommentsReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/posts/%d/comments?page=1&per_page=10", postID), nil)
	listCommentsResp, err := app.Test(listCommentsReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, listCommentsResp.StatusCode)

	getCommentReq := httptest.NewRequest(http.MethodGet, "/api/v1/comments/1", nil)
	getCommentResp, err := app.Test(getCommentReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, getCommentResp.StatusCode)

	updateCommentBody := []byte(`{"content":"Komentar sudah diupdate oleh pemiliknya!"}`)
	updateCommentReq := httptest.NewRequest(http.MethodPut, "/api/v1/comments/1", bytes.NewReader(updateCommentBody))
	updateCommentReq.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	updateCommentReq.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)
	updateCommentResp, err := app.Test(updateCommentReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, updateCommentResp.StatusCode)

	delCommentReq := httptest.NewRequest(http.MethodDelete, "/api/v1/comments/1", nil)
	delCommentReq.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)
	delCommentResp, err := app.Test(delCommentReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, delCommentResp.StatusCode)

	delPostReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/v1/posts/%d", postID), nil)
	delPostReq.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)
	delPostResp, err := app.Test(delPostReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, delPostResp.StatusCode)

	healthReq := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	healthResp, err := app.Test(healthReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, healthResp.StatusCode)

	readyReq := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	readyResp, err := app.Test(readyReq)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, readyResp.StatusCode)
}

