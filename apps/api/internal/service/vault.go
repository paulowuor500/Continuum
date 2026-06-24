package service

import (
	"context"
	"errors"

	"continuum/api/internal/model"
	"continuum/api/internal/repository"
)

type VaultService struct {
	repo      *repository.Database
	lightning *LightningService
}

func NewVaultService(repo *repository.Database, ln *LightningService) *VaultService {
	return &VaultService{repo: repo, lightning: ln}
}

func (s *VaultService) CreateNewVault(ctx context.Context, vault *model.Vault) error {
	if vault.CheckInIntervalSeconds <= 0 {
		return errors.New("check-in interval window must be greater than zero")
	}
	if vault.MultisigRequired <= 0 || len(vault.MultisigPubkeys) == 0 || vault.MultisigDescriptor == "" {
		return errors.New("missing multisig recovery policy")
	}
	return s.repo.InsertVault(ctx, vault)
}

// RequestCheckInInvoice generates a real 1-sat invoice bound to a vault security check-in
func (s *VaultService) RequestCheckInInvoice(ctx context.Context, vaultID string) (string, error) {
	// Confirm vault space exists before poking LND
	_, err := s.repo.GetVaultByID(ctx, vaultID)
	if err != nil {
		return "", errors.New("cannot generate check-in token: vault space not found")
	}

	// Route down to our LND engine layer to extract an invoice string
	invoice, err := s.lightning.GenerateProofInvoice(ctx, vaultID)
	if err != nil {
		return "", err
	}

	return invoice, nil
}

// VerifyAndProcessCheckIn completes the proof loop after invoice preimage payment confirmation
func (s *VaultService) VerifyAndProcessCheckIn(ctx context.Context, vaultID string, paymentPreimage string) error {
	if paymentPreimage == "" {
		return errors.New("cryptographic check-in proof requires a valid payment preimage")
	}

	// NOTE: Production LND validation hook goes here:
	// s.lightning.VerifyPaymentCleared(paymentPreimage)

	// Bumps last seen metrics back to NOW() and resets status to ACTIVE
	query := `
		UPDATE vaults 
		SET last_check_in_at = NOW(), 
		    status = 'ACTIVE',
		    updated_at = NOW()
		WHERE id = $1;
	`
	result, err := s.repo.Db.ExecContext(ctx, query, vaultID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("target continuum vault space not found")
	}

	return nil
}
