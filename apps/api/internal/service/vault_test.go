package service

import (
	"context"
	"errors"
	"testing"

	"continuum/api/internal/model"
)

func TestCreateNewVaultRejectsInvalidPolicy(t *testing.T) {
	svc := &VaultService{}
	vault := &model.Vault{
		Alias:                "Demo",
		BeneficiaryPubkey:    "02aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		EncryptedPayload:     "cipher",
		CheckInIntervalSeconds: 60,
		MultisigRequired:     2,
		MultisigPubkeys:      []string{"02aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		MultisigDescriptor:   "",
	}

	err := svc.CreateNewVault(context.Background(), vault)
	if err == nil {
		t.Fatal("expected validation error for missing multisig descriptor")
	}
}

func TestCreateNewVaultRejectsNonPositiveInterval(t *testing.T) {
	svc := &VaultService{}
	vault := &model.Vault{
		Alias:                "Demo",
		BeneficiaryPubkey:    "02aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		EncryptedPayload:     "cipher",
		CheckInIntervalSeconds: 0,
		MultisigRequired:     2,
		MultisigPubkeys:      []string{"02aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"},
		MultisigDescriptor:   "wsh(sortedmulti(2,02aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa))",
	}

	err := svc.CreateNewVault(context.Background(), vault)
	if err == nil {
		t.Fatal("expected validation error for non-positive interval")
	}
}

func TestVerifyAndProcessCheckInRejectsMissingPreimage(t *testing.T) {
	svc := &VaultService{}
	err := svc.VerifyAndProcessCheckIn(context.Background(), "test-vault", "")
	if !errors.Is(err, errors.New("cryptographic check-in proof requires a valid payment preimage")) {
		if err == nil || err.Error() != "cryptographic check-in proof requires a valid payment preimage" {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}
