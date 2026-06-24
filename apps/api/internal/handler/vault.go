package handler

import (
	"encoding/json"
	"time"

	"continuum/api/internal/model"
	"continuum/api/internal/repository"
	"continuum/api/internal/service"

	"github.com/gofiber/fiber/v3"
	"github.com/lib/pq"
)

type VaultHandler struct {
	Repo         *repository.Database
	VaultService *service.VaultService
	Multisig     *service.MultisigService
	Lightning    *service.LightningService
}

func NewVaultHandler(repo *repository.Database, vaultSvc *service.VaultService, multisig *service.MultisigService, lightning *service.LightningService) *VaultHandler {
	return &VaultHandler{
		Repo:         repo,
		VaultService: vaultSvc,
		Multisig:     multisig,
		Lightning:    lightning,
	}
}

// addBeneficiaryRequest is the expected body for POST /api/vaults/:id/beneficiaries
type addBeneficiaryRequest struct {
	Pubkey string `json:"pubkey"`
}

// ListVaults handles GET /api/vaults
func (h *VaultHandler) ListVaults(c fiber.Ctx) error {
	vaults, err := h.Repo.ListVaults(c.Context())
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{"vaults": vaults})
}

// CreateVault handles POST /api/vaults
func (h *VaultHandler) CreateVault(c fiber.Ctx) error {
	body := c.Body()
	if len(body) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "Empty request body"})
	}

	var req model.Vault
	if err := json.Unmarshal(body, &req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request payload format"})
	}

	if req.Alias == "" || req.BeneficiaryPubkey == "" || req.EncryptedPayload == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing required fields: alias, beneficiary_pubkey, encrypted_payload"})
	}
	if req.CheckInIntervalSeconds <= 0 {
		req.CheckInIntervalSeconds = 2592000 // 30 days default
	}

	pubkeys := []string(req.MultisigPubkeys)
	if len(pubkeys) == 0 {
		pubkeys = defaultDemoPubkeys(req.BeneficiaryPubkey)
	}

	policy, err := h.Multisig.BuildPolicy(c.Context(), req.MultisigRequired, pubkeys)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}
	req.MultisigRequired = policy.Required
	req.MultisigPubkeys = pq.StringArray(policy.Pubkeys)
	req.MultisigAddress = policy.Address
	req.MultisigRedeemScript = policy.RedeemScript
	req.MultisigDescriptor = policy.Descriptor
	req.MultisigNetwork = policy.Network

	req.Status = "ACTIVE"
	req.LastCheckInAt = time.Now()

	if err := h.VaultService.CreateNewVault(c.Context(), &req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(201).JSON(fiber.Map{
		"status":   "SUCCESS",
		"vault_id": req.ID,
		"message":  "Vault created and activated.",
		"multisig": fiber.Map{
			"required":      req.MultisigRequired,
			"pubkeys":       req.MultisigPubkeys,
			"address":       req.MultisigAddress,
			"redeem_script": req.MultisigRedeemScript,
			"descriptor":    req.MultisigDescriptor,
			"network":       req.MultisigNetwork,
		},
	})
}

// RequestCheckInToken handles POST /api/vaults/:id/invoice
func (h *VaultHandler) RequestCheckInToken(c fiber.Ctx) error {
	vaultID := c.Params("id")
	if vaultID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing vault ID"})
	}

	invoice, err := h.VaultService.RequestCheckInInvoice(c.Context(), vaultID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status":          "PENDING_PAYMENT",
		"payment_request": invoice,
		"value_sats":      1,
		"message":         "Pay this 1-sat invoice to prove life status.",
	})
}

// ConfirmCheckIn handles POST /api/vaults/:id/checkin
func (h *VaultHandler) ConfirmCheckIn(c fiber.Ctx) error {
	vaultID := c.Params("id")
	if vaultID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing vault ID"})
	}

	var payload struct {
		Preimage string `json:"preimage"`
	}
	if err := c.Bind().Body(&payload); err != nil || payload.Preimage == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing payment preimage"})
	}

	if err := h.VaultService.VerifyAndProcessCheckIn(c.Context(), vaultID, payload.Preimage); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Proof of life accepted. Vault timer reset.",
	})
}

// AddBeneficiary handles POST /api/vaults/:id/beneficiaries
func (h *VaultHandler) AddBeneficiary(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing vault ID"})
	}

	var req addBeneficiaryRequest
	if err := c.Bind().Body(&req); err != nil || req.Pubkey == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing beneficiary pubkey"})
	}

	vault, err := h.Repo.GetVaultByID(c.Context(), id)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Vault not found"})
	}

	pubkeys := append([]string{}, vault.MultisigPubkeys...)
	pubkeys = append(pubkeys, req.Pubkey)

	policy, err := h.Multisig.BuildPolicy(c.Context(), vault.MultisigRequired, pubkeys)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	vault.MultisigRequired = policy.Required
	vault.MultisigPubkeys = pq.StringArray(policy.Pubkeys)
	vault.MultisigAddress = policy.Address
	vault.MultisigRedeemScript = policy.RedeemScript
	vault.MultisigDescriptor = policy.Descriptor
	vault.MultisigNetwork = policy.Network

	if err := h.Repo.UpdateVaultMultisigPolicy(c.Context(), vault); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status":  "SUCCESS",
		"message": "Beneficiary added to multisig recovery policy.",
		"multisig": fiber.Map{
			"required":      vault.MultisigRequired,
			"pubkeys":       vault.MultisigPubkeys,
			"address":       vault.MultisigAddress,
			"redeem_script": vault.MultisigRedeemScript,
			"descriptor":    vault.MultisigDescriptor,
			"network":       vault.MultisigNetwork,
		},
	})
}

// CreateProofInvoice handles POST /api/vaults/:id/invoice (alias via Lightning directly)
func (h *VaultHandler) CreateProofInvoice(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Missing vault ID"})
	}

	invoice, err := h.Lightning.GenerateProofInvoice(c.Context(), id)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"status":  "SUCCESS",
		"invoice": invoice,
	})
}

func defaultDemoPubkeys(beneficiaryPubkey string) []string {
	return []string{
		beneficiaryPubkey,
		"02aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"03bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	}
}
