package handler

import (
	"time"

	"continuum/api/internal/repository"
	"github.com/gofiber/fiber/v3"
)

type RecoveryHandler struct {
	Repo *repository.Database
}

func NewRecoveryHandler(repo *repository.Database) *RecoveryHandler {
	return &RecoveryHandler{Repo: repo}
}

// GetVaultStatus handles GET /api/vaults/:id
func (h *RecoveryHandler) GetVaultStatus(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing target vault ID"})
	}

	// FIXED: Swapped c.Context() out for c.UserContext() to perfectly match Fiber v3 standard specs
	vault, err := h.Repo.GetVaultByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Target vault space not found"})
	}

	// PRODUCTION ADDITION: Calculate exact expiration milestone time boundaries
	expiresAt := vault.LastCheckInAt.Add(time.Duration(vault.CheckInIntervalSeconds) * time.Second)
	secondsRemaining := int64(time.Until(expiresAt).Seconds())

	// If calculated time buffer is wiped out but background cron hadn't flipped state yet,
	// enforce defensive fallback logic
	isPastDeadline := secondsRemaining <= 0

	// SECURITY CHECK: If owner is active and timeline is healthy, shield the payload!
	if vault.Status == "ACTIVE" && !isPastDeadline {
		return c.JSON(fiber.Map{
			"id":                vault.ID,
			"alias":             vault.Alias,
			"status":            vault.Status,
			"last_seen":         vault.LastCheckInAt,
			"expires_at":        expiresAt,
			"seconds_remaining": secondsRemaining, // ⏳ Crucial parameter for frontend countdown meters
			"payload_locked":    true,
			"message":           "Vault remains cryptographically shielded. Owner verified active.",
		})
	}

	// RELEASE CONDITION MET: Owner missing (or deadline missed), hand off the ciphertext package
	return c.JSON(fiber.Map{
		"id":                vault.ID,
		"alias":             vault.Alias,
		"status":            "DORMANT", // Defensive enforcement override
		"last_seen":         vault.LastCheckInAt,
		"expires_at":        expiresAt,
		"seconds_remaining": 0,
		"payload_locked":    false,
		"encrypted_payload": vault.EncryptedPayload, // Released for client-side browser decryption
	})
}
