package repository

import (
    "testing"

    "continuum/api/internal/model"
)

func TestNormalizeVaultArraysInitializesNilSlice(t *testing.T) {
    vault := &model.Vault{}

    NormalizeVaultArrays(vault)

    if vault.MultisigPubkeys == nil {
        t.Fatal("expected multisig pubkeys slice to be initialized")
    }
}
