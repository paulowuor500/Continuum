# Continuum Architecture

## Overview

Continuum is a Bitcoin inheritance protocol that uses Lightning Network proof-of-life payments to determine wallet activity and trigger inheritance recovery workflows when a user becomes inactive.

---

## Components

### Web Application

Provides the user interface for:

- Creating inheritance vaults
- Managing beneficiaries
- Uploading encrypted recovery packages
- Viewing vault status

### API Service

Built with Go and Fiber.

Responsibilities:

- Vault creation
- Vault status retrieval
- Recovery workflows
- Lightning proof-of-life integration

---

## Core Services

### Vault Service

Stores inheritance vault metadata including:

- Alias
- Beneficiary public key
- Encrypted recovery package

### Lightning Service

Generates 1-satoshi Lightning invoices.

Payment of an invoice acts as a proof-of-life signal.

### Scheduler Service

Runs periodic inactivity checks.

If:

last_check_in_at + check_in_interval < current_time

Then:

ACTIVE → DORMANT

### Recovery Service

Controls release of encrypted recovery packages.

ACTIVE vaults:
- Payload remains locked.

DORMANT vaults:
- Payload becomes accessible for beneficiary recovery.

---

## Recovery Flow

User
↓
Creates Vault
↓
Uploads Encrypted Recovery Package
↓
Periodically Pays Lightning Proof-of-Life Invoice
↓
Scheduler Monitors Activity
↓
No Check-In Detected
↓
Vault Marked DORMANT
↓
Encrypted Recovery Package Released

---

## Testing Utilities

### Time Warp Endpoint

POST /api/vaults/:id/warp

Simulates long-term inactivity by moving the last check-in timestamp 100 days into the past.
