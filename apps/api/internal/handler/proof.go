package handler

import (
	"continuum/api/internal/repository"
	"continuum/api/internal/service"
	"github.com/gofiber/fiber/v3"
)

type ProofHandler struct {
	Repo             *repository.Database
	LightningService *service.LightningService 
}

// NewProofHandler now accepts both the relational storage pool and the lightning engine
func NewProofHandler(repo *repository.Database, lnService *service.LightningService) *ProofHandler {
	return &ProofHandler{
		Repo:             repo,
		LightningService: lnService,
	}
}

// SimulateTimeWarp handles POST /api/vaults/:id/warp
func (h *ProofHandler) SimulateTimeWarp(c fiber.Ctx) error {
	vaultID := c.Params("id")
	if vaultID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing target vault ID"})
	}

	// Forcibly set the check-in time back 100 days into the past
	query := `
		UPDATE vaults 
		SET last_check_in_at = NOW() - INTERVAL '100 days' 
		WHERE id = $1;
	`
	result, err := h.Repo.Db.ExecContext(c.Context(), query, vaultID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	if rowsAffected == 0 {
		return c.Status(404).JSON(fiber.Map{"error": "Target vault space not found"})
	}

	return c.JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Time warped 100 days back. Background loop will flag vault DORMANT on next tick.",
	})
}
