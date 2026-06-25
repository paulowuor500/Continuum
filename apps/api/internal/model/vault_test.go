package model

import (
	"testing"
	"time"
)

func TestVaultDefaultsAndStatusValues(t *testing.T) {
	vault := Vault{
		Alias:                "Test Vault",
		BeneficiaryPubkey:    "pubkey",
		EncryptedPayload:     "payload",
		CheckInIntervalSeconds: 60,
		LastCheckInAt:        time.Now(),
		Status:               StatusActive,
	}

	if vault.Alias != "Test Vault" {
		t.Fatalf("expected alias to be preserved, got %q", vault.Alias)
	}
	if vault.Status != StatusActive {
		t.Fatalf("expected active status, got %q", vault.Status)
	}
	if vault.CheckInIntervalSeconds != 60 {
		t.Fatalf("expected interval to be preserved, got %d", vault.CheckInIntervalSeconds)
	}
}

func TestVaultStatusConstants(t *testing.T) {
	if StatusActive == "" || StatusDormant == "" {
		t.Fatal("expected status constants to be populated")
	}
	if StatusActive == StatusDormant {
		t.Fatal("expected active and dormant statuses to differ")
	}
}
