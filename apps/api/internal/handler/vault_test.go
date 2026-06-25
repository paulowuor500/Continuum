package handler

import (
    "net/http"
    "testing"

    "github.com/gofiber/fiber/v3"
)

func TestDefaultDemoPubkeysIncludesBeneficiaryAndOperators(t *testing.T) {
    pubkeys := defaultDemoPubkeys("02abc")
    if len(pubkeys) != 3 {
        t.Fatalf("expected 3 demo pubkeys, got %d", len(pubkeys))
    }
    if pubkeys[0] != "02abc" {
        t.Fatalf("expected beneficiary key first, got %s", pubkeys[0])
    }
}

func TestCreateVaultRejectsEmptyBody(t *testing.T) {
    app := fiber.New()
    h := &VaultHandler{}

    app.Post("/vaults", h.CreateVault)

    req, _ := http.NewRequest(http.MethodPost, "/vaults", nil)
    resp, err := app.Test(req)
    if err != nil {
        t.Fatalf("request failed: %v", err)
    }
    if resp.StatusCode != fiber.StatusBadRequest {
        t.Fatalf("expected 400, got %d", resp.StatusCode)
    }
}
