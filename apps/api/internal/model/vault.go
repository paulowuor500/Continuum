package model

import (
	"time"

	"github.com/lib/pq"
)

type VaultStatus string

const (
	StatusActive  VaultStatus = "ACTIVE"
	StatusDormant VaultStatus = "DORMANT"
)

type Vault struct {
	// omitempty ensures creation payloads don't override DB-generated UUID fields.
	ID                   string         `db:"id"                      json:"id,omitempty"`
	Alias                string         `db:"alias"                   json:"alias"`
	BeneficiaryPubkey    string         `db:"beneficiary_pubkey"      json:"beneficiary_pubkey"`
	EncryptedPayload     string         `db:"encrypted_payload"       json:"encrypted_payload"`
	CheckInIntervalSeconds int          `db:"check_in_interval_seconds" json:"check_in_interval_seconds"`
	LastCheckInAt        time.Time      `db:"last_check_in_at"        json:"last_check_in_at"`
	Status               VaultStatus    `db:"status"                  json:"status,omitempty"`

	// Multisig recovery policy fields
	MultisigRequired     int            `db:"multisig_required"       json:"multisig_required"`
	MultisigPubkeys      pq.StringArray `db:"multisig_pubkeys"        json:"multisig_pubkeys"`
	MultisigAddress      string         `db:"multisig_address"        json:"multisig_address"`
	MultisigRedeemScript string         `db:"multisig_redeem_script"  json:"multisig_redeem_script"`
	MultisigDescriptor   string         `db:"multisig_descriptor"     json:"multisig_descriptor"`
	MultisigNetwork      string         `db:"multisig_network"        json:"multisig_network"`

	// Audit timestamps
	CreatedAt            time.Time      `db:"created_at"              json:"created_at,omitempty"`
	UpdatedAt            time.Time      `db:"updated_at"              json:"updated_at,omitempty"`
}
