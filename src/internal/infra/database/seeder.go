package database

import (
	"context"
	"fmt"
	"log/slog"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"maj-core-srvc/src/internal/model"
)

// SeedAll mengisi data awal (users, posts, comments) secara idempotent di dalam satu transaksi database.
func SeedAll(ctx context.Context, db *gorm.DB) error {
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		users, err := seedUsers(tx)
		if err != nil {
			return fmt.Errorf("seed users: %w", err)
		}

		posts, err := seedPosts(tx, users)
		if err != nil {
			return fmt.Errorf("seed posts: %w", err)
		}

		if err := seedComments(tx, users, posts); err != nil {
			return fmt.Errorf("seed comments: %w", err)
		}

		return nil
	})
}

func seedUsers(tx *gorm.DB) (map[string]*model.User, error) {
	defaultPassHash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	seeds := []model.User{
		{
			Name:     "Admin Majoo",
			Email:    "admin@majoo.id",
			Password: string(defaultPassHash),
			Role:     model.RoleAdmin,
		},
		{
			Name:     "Budi Penulis",
			Email:    "author@majoo.id",
			Password: string(defaultPassHash),
			Role:     model.RoleUser,
		},
		{
			Name:     "Siti Pembaca",
			Email:    "reader@majoo.id",
			Password: string(defaultPassHash),
			Role:     model.RoleUser,
		},
	}

	result := make(map[string]*model.User, len(seeds))
	for i := range seeds {
		u := seeds[i]
		var existing model.User
		err := tx.Where("email = ?", u.Email).First(&existing).Error
		if err == nil {
			result[u.Email] = &existing
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}

		if err := tx.Create(&u).Error; err != nil {
			return nil, err
		}
		slog.Info("seeded user", "id", u.ID, "email", u.Email, "role", u.Role)
		result[u.Email] = &u
	}

	return result, nil
}

func seedPosts(tx *gorm.DB, users map[string]*model.User) (map[string]*model.Post, error) {
	author := users["author@majoo.id"]
	admin := users["admin@majoo.id"]

	seeds := []model.Post{
		{
			AuthorID:    author.ID,
			Title:       "Membangun Microservice Golang dengan Fiber dan GORM",
			Slug:        "membangun-microservice-golang-dengan-fiber-dan-gorm",
			Content:     "Panduan lengkap merancang boilerplate microservice yang modular, cepat, dan mudah di-maintain menggunakan Fiber dan GORM.",
			IsPublished: true,
		},
		{
			AuthorID:    author.ID,
			Title:       "Pentingnya Database Transaction pada Sistem Berskala Besar",
			Slug:        "pentingnya-database-transaction-pada-sistem-berskala-besar",
			Content:     "Bagaimana menjaga konsistensi data antar tabel menggunakan pola TxManager dan context propagation di Golang.",
			IsPublished: true,
		},
		{
			AuthorID:    admin.ID,
			Title:       "Standarisasi HTTP Response dan Error Handling",
			Slug:        "standarisasi-http-response-dan-error-handling",
			Content:     "Menyatukan format JSON response dan mapping kode error agar komunikasi antar service menjadi lebih prediktif.",
			IsPublished: true,
		},
	}

	result := make(map[string]*model.Post, len(seeds))
	for i := range seeds {
		p := seeds[i]
		var existing model.Post
		err := tx.Where("slug = ?", p.Slug).First(&existing).Error
		if err == nil {
			result[p.Slug] = &existing
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return nil, err
		}

		if err := tx.Create(&p).Error; err != nil {
			return nil, err
		}
		slog.Info("seeded post", "id", p.ID, "slug", p.Slug)
		result[p.Slug] = &p
	}

	return result, nil
}

func seedComments(tx *gorm.DB, users map[string]*model.User, posts map[string]*model.Post) error {
	reader := users["reader@majoo.id"]
	author := users["author@majoo.id"]
	admin := users["admin@majoo.id"]

	post1 := posts["membangun-microservice-golang-dengan-fiber-dan-gorm"]
	post2 := posts["pentingnya-database-transaction-pada-sistem-berskala-besar"]

	type commentSeed struct {
		Post    *model.Post
		Author  *model.User
		Content string
	}

	seeds := []commentSeed{
		{
			Post:    post1,
			Author:  reader,
			Content: "Artikelnya sangat jelas! Struktur 3-tier dengan pemisahan DTO ini bikin kode jauh lebih rapi.",
		},
		{
			Post:    post1,
			Author:  admin,
			Content: "Setuju, jangan lupa selalu jalankan unit test sebelum merge PR ya.",
		},
		{
			Post:    post2,
			Author:  author,
			Content: "Semoga membantu teman-teman yang sedang belajar transaksi database di Go!",
		},
	}

	for _, cs := range seeds {
		if cs.Post == nil || cs.Author == nil {
			continue
		}

		var count int64
		err := tx.Model(&model.Comment{}).
			Where("post_id = ? AND author_id = ? AND content = ?", cs.Post.ID, cs.Author.ID, cs.Content).
			Count(&count).Error
		if err != nil {
			return err
		}
		if count > 0 {
			continue
		}

		comment := model.Comment{
			PostID:   cs.Post.ID,
			AuthorID: cs.Author.ID,
			Content:  cs.Content,
		}
		if err := tx.Create(&comment).Error; err != nil {
			return err
		}

		if err := tx.Model(&model.Post{}).
			Where("id = ?", cs.Post.ID).
			UpdateColumn("comments_count", gorm.Expr("comments_count + 1")).Error; err != nil {
			return err
		}

		slog.Info("seeded comment", "id", comment.ID, "post_id", cs.Post.ID, "author_id", cs.Author.ID)
	}

	return nil
}
