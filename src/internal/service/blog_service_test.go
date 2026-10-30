package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"maj-core-srvc/src/internal/config"
	servicedto "maj-core-srvc/src/internal/dto/service"
	"maj-core-srvc/src/internal/infra/database"
	"maj-core-srvc/src/internal/model"
	"maj-core-srvc/src/internal/pkg/response"
	"maj-core-srvc/src/internal/repository"
	"maj-core-srvc/src/internal/service"
)

func setupTestBlogDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&model.User{}, &model.Post{}, &model.Comment{})
	require.NoError(t, err)
	return db
}

func TestBlogFlow_Auth_CRUD_And_Transactions(t *testing.T) {
	ctx := context.Background()
	db := setupTestBlogDB(t)

	txManager := database.NewTxManager(db)
	userRepo := repository.NewUserRepository(db)
	postRepo := repository.NewPostRepository(db)
	commentRepo := repository.NewCommentRepository(db)

	authSvc := service.NewAuthService(userRepo, config.AuthConfig{
		JWTSecret:     "test-secret",
		JWTIssuer:     "test-issuer",
		JWTExpiration: time.Hour,
	})
	postSvc := service.NewPostService(txManager, postRepo, commentRepo)
	commentSvc := service.NewCommentService(txManager, postRepo, commentRepo)

	// 1. Register Author, User biasa, dan Admin
	authorAuth, err := authSvc.Register(ctx, servicedto.RegisterUserInput{
		Name:     "Penulis Utama",
		Email:    "author@majoo.id",
		Password: "password123",
		Role:     model.RoleUser,
	})
	require.NoError(t, err)
	require.NotEmpty(t, authorAuth.AccessToken)
	authorID := authorAuth.User.ID

	otherUserAuth, err := authSvc.Register(ctx, servicedto.RegisterUserInput{
		Name:     "Pembaca",
		Email:    "reader@majoo.id",
		Password: "password123",
		Role:     model.RoleUser,
	})
	require.NoError(t, err)
	otherUserID := otherUserAuth.User.ID

	adminAuth, err := authSvc.Register(ctx, servicedto.RegisterUserInput{
		Name:     "Moderator Admin",
		Email:    "admin@majoo.id",
		Password: "password123",
		Role:     model.RoleAdmin,
	})
	require.NoError(t, err)
	adminID := adminAuth.User.ID

	// Duplicate email harus ditolak (409 DUPLICATE_ENTRY)
	_, err = authSvc.Register(ctx, servicedto.RegisterUserInput{
		Name:     "Duplikat",
		Email:    "author@majoo.id",
		Password: "password123",
	})
	require.Error(t, err)
	var appErr *response.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, response.CodeDuplicateEntry, appErr.Code)

	// Login dengan password salah harus ditolak (401 UNAUTHORIZED)
	_, err = authSvc.Login(ctx, servicedto.LoginInput{
		Email:    "author@majoo.id",
		Password: "wrong-password",
	})
	require.Error(t, err)
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, response.CodeUnauthorized, appErr.Code)

	// Login dengan password benar harus berhasil & GetProfile mengembalikan data user
	loginRes, err := authSvc.Login(ctx, servicedto.LoginInput{
		Email:    "author@majoo.id",
		Password: "password123",
	})
	require.NoError(t, err)
	assert.NotEmpty(t, loginRes.AccessToken)

	profileRes, err := authSvc.GetProfile(ctx, authorID)
	require.NoError(t, err)
	assert.Equal(t, "author@majoo.id", profileRes.Email)

	// 2. Author membuat Post baru
	post, err := postSvc.Create(ctx, servicedto.CreatePostInput{
		AuthorID:    authorID,
		Title:       "Arsitektur Microservice di Golang",
		Content:     "Ini adalah artikel mendalam tentang pola 3-tier dan transaksi database.",
		IsPublished: true,
	})
	require.NoError(t, err)
	assert.Equal(t, authorID, post.AuthorID)
	assert.Equal(t, 0, post.CommentsCount)

	// List Posts dengan filter pencarian & author_id
	postsList, totalPosts, err := postSvc.List(ctx, servicedto.ListPostFilter{
		Page:     1,
		PerPage:  10,
		Search:   "Microservice",
		AuthorID: authorID,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), totalPosts)
	assert.Len(t, postsList, 1)

	// Author mengupdate Post miliknya sendiri -> 200 OK
	updatedOwnPost, err := postSvc.Update(ctx, servicedto.UpdatePostInput{
		ID:          post.ID,
		ActorID:     authorID,
		ActorRole:   model.RoleUser,
		Title:       "Arsitektur Microservice di Golang (Updated)",
		Content:     "Konten artikel yang sudah diperbarui oleh penulis aslinya.",
		IsPublished: true,
	})
	require.NoError(t, err)
	assert.Equal(t, "Arsitektur Microservice di Golang (Updated)", updatedOwnPost.Title)

	// 3. Authorization Check: User lain TIDAK BOLEH mengedit post milik Author (403 FORBIDDEN)
	_, err = postSvc.Update(ctx, servicedto.UpdatePostInput{
		ID:          post.ID,
		ActorID:     otherUserID,
		ActorRole:   model.RoleUser,
		Title:       "Judul Bajakan",
		Content:     "Konten bajakan minimal 10 karakter.",
		IsPublished: true,
	})
	require.Error(t, err)
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, response.CodeForbidden, appErr.Code)

	// 4. Transactional Comment Creation: Pembaca menambahkan 2 komentar -> comments_count pada Post otomatis naik jadi 2
	c1, err := commentSvc.Create(ctx, servicedto.CreateCommentInput{
		PostID:   post.ID,
		AuthorID: otherUserID,
		Content:  "Artikel yang sangat bermanfaat!",
	})
	require.NoError(t, err)

	c2, err := commentSvc.Create(ctx, servicedto.CreateCommentInput{
		PostID:   post.ID,
		AuthorID: authorID,
		Content:  "Terima kasih sudah mampir!",
	})
	require.NoError(t, err)

	updatedPost, err := postSvc.GetByID(ctx, post.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, updatedPost.CommentsCount)

	// GetByID, ListByPostID, dan Update Comment oleh pemiliknya sendiri
	c1Detail, err := commentSvc.GetByID(ctx, c1.ID)
	require.NoError(t, err)
	assert.Equal(t, c1.ID, c1Detail.ID)

	commentsList, totalComments, err := commentSvc.ListByPostID(ctx, servicedto.ListCommentFilter{
		PostID:  post.ID,
		Page:    1,
		PerPage: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, int64(2), totalComments)
	assert.Len(t, commentsList, 2)

	updatedC1, err := commentSvc.Update(ctx, servicedto.UpdateCommentInput{
		ID:        c1.ID,
		ActorID:   otherUserID,
		ActorRole: model.RoleUser,
		Content:   "Artikel yang sangat bermanfaat! (Edited)",
	})
	require.NoError(t, err)
	assert.Equal(t, "Artikel yang sangat bermanfaat! (Edited)", updatedC1.Content)

	// 5. Anti-IDOR Check: User lain (termasuk Admin sekalipun) TIDAK BOLEH mengedit atau menghapus komentar/post milik orang lain (403 FORBIDDEN)
	_, err = commentSvc.Update(ctx, servicedto.UpdateCommentInput{
		ID:        c1.ID,
		ActorID:   authorID,
		ActorRole: model.RoleUser,
		Content:   "Coba bajak komentar orang lain",
	})
	require.Error(t, err)
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, response.CodeForbidden, appErr.Code)

	err = commentSvc.Delete(ctx, servicedto.DeleteCommentInput{
		ID:        c1.ID,
		ActorID:   adminID,
		ActorRole: model.RoleAdmin,
	})
	require.Error(t, err)
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, response.CodeForbidden, appErr.Code)

	err = postSvc.Delete(ctx, servicedto.DeletePostInput{
		ID:        post.ID,
		ActorID:   otherUserID,
		ActorRole: model.RoleUser,
	})
	require.Error(t, err)
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, response.CodeForbidden, appErr.Code)

	// 6. Transactional Comment Deletion oleh pemilik komentar sendiri (otherUserID) -> comments_count berkurang menjadi 1
	err = commentSvc.Delete(ctx, servicedto.DeleteCommentInput{
		ID:        c1.ID,
		ActorID:   otherUserID,
		ActorRole: model.RoleUser,
	})
	require.NoError(t, err)

	updatedPost, err = postSvc.GetByID(ctx, post.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, updatedPost.CommentsCount)

	// 7. Transactional Post Deletion oleh pemilik post sendiri (authorID): Hapus post sekaligus menghapus sisa komentar (c2) di dalam 1 transaksi DB
	err = postSvc.Delete(ctx, servicedto.DeletePostInput{
		ID:        post.ID,
		ActorID:   authorID,
		ActorRole: model.RoleUser,
	})
	require.NoError(t, err)

	// Pastikan post sudah terhapus (404 NOT_FOUND) dan komentar c2 juga ikut terhapus
	_, err = postSvc.GetByID(ctx, post.ID)
	require.Error(t, err)
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, response.CodeNotFound, appErr.Code)

	remainingComment, err := commentRepo.FindByID(ctx, c2.ID)
	require.NoError(t, err)
	assert.Nil(t, remainingComment)
}

