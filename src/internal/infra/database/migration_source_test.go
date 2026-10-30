package database_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"maj-core-srvc/src/internal/infra/database"
	"maj-core-srvc/src/internal/model"
)

func TestSingleFileSource_ParseAndRead(t *testing.T) {
	tmpDir := t.TempDir()
	migPath := filepath.Join(tmpDir, "20260101000001_create_users.sql")

	content := `-- +migrate Up
CREATE TABLE users (id BIGSERIAL PRIMARY KEY);

-- +migrate Down
DROP TABLE IF EXISTS users;
`
	require.NoError(t, os.WriteFile(migPath, []byte(content), 0o644))

	drv, err := database.NewSingleFileSource(tmpDir)
	require.NoError(t, err)
	defer func() { _ = drv.Close() }()

	firstVer, err := drv.First()
	require.NoError(t, err)
	assert.Equal(t, uint(20260101000001), firstVer)

	upReader, ident, err := drv.ReadUp(firstVer)
	require.NoError(t, err)
	defer func() { _ = upReader.Close() }()
	assert.Equal(t, "create_users", ident)

	upBytes, err := io.ReadAll(upReader)
	require.NoError(t, err)
	assert.Equal(t, "CREATE TABLE users (id BIGSERIAL PRIMARY KEY);", string(upBytes))

	downReader, ident, err := drv.ReadDown(firstVer)
	require.NoError(t, err)
	defer func() { _ = downReader.Close() }()
	assert.Equal(t, "create_users", ident)

	downBytes, err := io.ReadAll(downReader)
	require.NoError(t, err)
	assert.Equal(t, "DROP TABLE IF EXISTS users;", string(downBytes))
}

func TestParseSingleFileSQL_MissingMarker(t *testing.T) {
	_, _, err := database.ParseSingleFileSQL("CREATE TABLE test (id INT);")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing '-- +migrate Up' marker")
}

func TestRepositoryMigrationsDirectory_PerTableOrder(t *testing.T) {
	drv, err := database.NewSingleFileSource("../../../../migrations")
	require.NoError(t, err)
	defer func() { _ = drv.Close() }()

	v1, err := drv.First()
	require.NoError(t, err)
	assert.Equal(t, uint(20260101000001), v1)

	v2, err := drv.Next(v1)
	require.NoError(t, err)
	assert.Equal(t, uint(20260101000002), v2)

	v3, err := drv.Next(v2)
	require.NoError(t, err)
	assert.Equal(t, uint(20260101000003), v3)
}

func TestSeedAll_Idempotent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Post{}, &model.Comment{}))

	ctx := context.Background()

	// Run 1: harus mengisi 3 users, 3 posts, dan 3 comments
	require.NoError(t, database.SeedAll(ctx, db))

	// Run 2 (idempotent): dijalankan ulang tidak boleh error atau menduplikasi data
	require.NoError(t, database.SeedAll(ctx, db))

	var userCount, postCount, commentCount int64
	require.NoError(t, db.Model(&model.User{}).Count(&userCount).Error)
	require.NoError(t, db.Model(&model.Post{}).Count(&postCount).Error)
	require.NoError(t, db.Model(&model.Comment{}).Count(&commentCount).Error)

	assert.Equal(t, int64(3), userCount)
	assert.Equal(t, int64(3), postCount)
	assert.Equal(t, int64(3), commentCount)
}

