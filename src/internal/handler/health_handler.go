package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"maj-core-srvc/src/internal/infra/database"
	"maj-core-srvc/src/internal/pkg/response"
)

type HealthHandler struct {
	db *gorm.DB
}

func NewHealthHandler(db *gorm.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

// Liveness memeriksa apakah proses HTTP server masih berjalan.
//
// @Summary      Liveness Probe
// @Description  Memeriksa apakah proses HTTP server masih hidup
// @Tags         Health
// @Produce      json
// @Success      200  {object}  response.Envelope
// @Router       /healthz [get]
func (h *HealthHandler) Liveness(c *fiber.Ctx) error {
	return response.OK(c, "Service is alive", fiber.Map{
		"status": "UP",
	})
}

// Readiness memeriksa apakah connection pool database sehat dan siap melayani traffic.
//
// @Summary      Readiness Probe
// @Description  Memeriksa koneksi database dan menampilkan statistik connection pool
// @Tags         Health
// @Produce      json
// @Success      200  {object}  response.Envelope
// @Failure      503  {object}  response.Envelope
// @Router       /readyz [get]
func (h *HealthHandler) Readiness(c *fiber.Ctx) error {
	sqlDB, err := h.db.DB()
	if err != nil {
		return response.WrapError(response.CodeServiceUnavailable, "Database instance unavailable", err)
	}

	ctx, cancel := context.WithTimeout(c.UserContext(), 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return response.WrapError(response.CodeServiceUnavailable, "Database ping failed", err)
	}

	poolStats, _ := database.Stats(h.db)

	return response.OK(c, "Service is ready", fiber.Map{
		"status":   "READY",
		"database": "CONNECTED",
		"pool":     poolStats,
	})
}
