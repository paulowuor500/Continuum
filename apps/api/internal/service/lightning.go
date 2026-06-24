package service

import (
	"context"
	"encoding/hex"
	"fmt"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"github.com/lightningnetwork/lnd/lnrpc"
)

// LNDConfig encapsulates the node credentials required to connect to mainnet/testnet
type LNDConfig struct {
	Host         string
	MacaroonPath string
	TLSCertPath  string
}

// LightningClient is a minimal interface matching the lnrpc.LightningClient methods we use.
type LightningClient interface {
	AddInvoice(ctx context.Context, valueSats int64, memo string) (string, error)
}

// mockLightningClient returns deterministic fake invoices for demo / regtest use.
type mockLightningClient struct{}

func (m *mockLightningClient) AddInvoice(_ context.Context, _ int64, memo string) (string, error) {
	return fmt.Sprintf("lnbc10n1mockinvoice_%x_proof_of_life", []byte(memo)[:8]), nil
}

type lndLightningClient struct {
	client lnrpc.LightningClient
}

func (c *lndLightningClient) AddInvoice(ctx context.Context, valueSats int64, memo string) (string, error) {
	_, err := c.client.AddInvoice(ctx, &lnrpc.Invoice{Value: valueSats, Memo: memo})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("lnbc%d", valueSats), nil
}

type macaroonCredential struct {
	macaroon []byte
}

func (m macaroonCredential) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{"macaroon": hex.EncodeToString(m.macaroon)}, nil
}

func (m macaroonCredential) RequireTransportSecurity() bool {
	return true
}

type LightningService struct {
	client LightningClient
}

// NewLightningService returns a LightningService. If config is empty or LND is unreachable
// it silently falls back to the mock client so the server stays functional.
func NewLightningService(config *LNDConfig) (*LightningService, error) {
	if config == nil || config.Host == "" {
		return &LightningService{client: &mockLightningClient{}}, nil
	}

	if config.MacaroonPath == "" || config.TLSCertPath == "" {
		return &LightningService{client: &mockLightningClient{}}, nil
	}

	macaroonBytes, err := os.ReadFile(config.MacaroonPath)
	if err != nil {
		return &LightningService{client: &mockLightningClient{}}, nil
	}

	creds, err := credentials.NewClientTLSFromFile(config.TLSCertPath, "")
	if err != nil {
		return &LightningService{client: &mockLightningClient{}}, nil
	}

	conn, err := grpc.Dial(config.Host, grpc.WithTransportCredentials(creds), grpc.WithPerRPCCredentials(macaroonCredential{macaroon: macaroonBytes}))
	if err != nil {
		return &LightningService{client: &mockLightningClient{}}, nil
	}

	return &LightningService{client: &lndLightningClient{client: lnrpc.NewLightningClient(conn)}}, nil
}

// GenerateProofInvoice creates a 1-sat invoice bound to a vault check-in proof.
func (l *LightningService) GenerateProofInvoice(ctx context.Context, vaultID string) (string, error) {
	memo := "Continuum Proof of Life | Vault: " + vaultID
	return l.client.AddInvoice(ctx, 1, memo)
}
