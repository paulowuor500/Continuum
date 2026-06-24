package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

type MultisigPolicy struct {
	Required     int
	Pubkeys      []string
	Address      string
	RedeemScript string
	Descriptor   string
	Network      string
}

type BitcoinRPCClient struct {
	endpoint string
	user     string
	password string
	client   *http.Client
}

type MultisigService struct {
	rpc *BitcoinRPCClient
}

var compressedPubkeyPattern = regexp.MustCompile(`^(02|03)[0-9a-fA-F]{64}$`)

func NewMultisigServiceFromEnv() *MultisigService {
	endpoint := os.Getenv("BITCOIN_RPC_URL")
	if endpoint == "" {
		endpoint = "http://127.0.0.1:18443"
	}

	user := os.Getenv("BITCOIN_RPC_USER")
	if user == "" {
		user = "continuum"
	}

	password := os.Getenv("BITCOIN_RPC_PASSWORD")
	if password == "" {
		password = "finality"
	}

	return &MultisigService{
		rpc: &BitcoinRPCClient{
			endpoint: endpoint,
			user:     user,
			password: password,
			client: &http.Client{
				Timeout: 2 * time.Second,
			},
		},
	}
}

func (s *MultisigService) BuildPolicy(ctx context.Context, required int, pubkeys []string) (*MultisigPolicy, error) {
	normalized, err := normalizePubkeys(pubkeys)
	if err != nil {
		return nil, err
	}

	if required <= 0 {
		required = 2
	}
	if required > len(normalized) {
		return nil, fmt.Errorf("multisig quorum %d exceeds signer count %d", required, len(normalized))
	}

	policy := &MultisigPolicy{
		Required:   required,
		Pubkeys:    normalized,
		Descriptor: descriptorFor(required, normalized),
		Network:    "regtest",
	}

	if s == nil || s.rpc == nil {
		return policy, nil
	}

	rpcPolicy, err := s.rpc.CreateMultisig(ctx, required, normalized)
	if err != nil {
		policy.Address = "offline-regtest-preview"
		return policy, nil
	}

	policy.Address = rpcPolicy.Address
	policy.RedeemScript = rpcPolicy.RedeemScript
	if rpcPolicy.Descriptor != "" {
		policy.Descriptor = rpcPolicy.Descriptor
	}
	return policy, nil
}

type rpcRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      string        `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

type createMultisigResult struct {
	Address      string   `json:"address"`
	RedeemScript string   `json:"redeemScript"`
	Descriptor   string   `json:"descriptor"`
	Warnings     []string `json:"warnings"`
}

type rpcResponse struct {
	Result createMultisigResult `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (c *BitcoinRPCClient) CreateMultisig(ctx context.Context, required int, pubkeys []string) (*MultisigPolicy, error) {
	payload := rpcRequest{
		JSONRPC: "1.0",
		ID:      "continuum-createmultisig",
		Method:  "createmultisig",
		Params:  []interface{}{required, pubkeys, "bech32"},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.user, c.password)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var decoded rpcResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, err
	}
	if decoded.Error != nil {
		return nil, errors.New(decoded.Error.Message)
	}

	return &MultisigPolicy{
		Required:     required,
		Pubkeys:      pubkeys,
		Address:      decoded.Result.Address,
		RedeemScript: decoded.Result.RedeemScript,
		Descriptor:   decoded.Result.Descriptor,
		Network:      "regtest",
	}, nil
}

func normalizePubkeys(pubkeys []string) ([]string, error) {
	seen := map[string]bool{}
	normalized := make([]string, 0, len(pubkeys))

	for _, key := range pubkeys {
		key = strings.ToLower(strings.TrimSpace(key))
		if key == "" {
			continue
		}
		if !compressedPubkeyPattern.MatchString(key) {
			return nil, fmt.Errorf("invalid compressed secp256k1 public key: %s", key)
		}
		if !seen[key] {
			seen[key] = true
			normalized = append(normalized, key)
		}
	}

	if len(normalized) < 2 {
		return nil, errors.New("multisig requires at least two unique compressed public keys")
	}

	sort.Strings(normalized)
	return normalized, nil
}

func descriptorFor(required int, pubkeys []string) string {
	return fmt.Sprintf("wsh(sortedmulti(%d,%s))", required, strings.Join(pubkeys, ","))
}
