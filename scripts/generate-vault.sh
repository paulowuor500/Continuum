#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

echo "🔐 Generating automated sample cryptographic capsule..."

# 1. Dynamically fetch a real public key from your running LND dev node (or fallback to mock)
if command -v docker &> /dev/null && docker ps | grep -q continuum-lnd; then
    echo "⚡ Fetching live public key from continuum-lnd node..."
    LIVE_PUBKEY=$(docker exec continuum-lnd lncli --network=regtest getinfo | grep "identity_pubkey" | cut -d '"' -f 4)
    TARGET_PUBKEY=${LIVE_PUBKEY:-"02e9a2631247d5124b893a71b25076eefc432d56a29851a7eef1109bcfa0329a1d"}
else
    TARGET_PUBKEY="02e9a2631247d5124b893a71b25076eefc432d56a29851a7eef1109bcfa0329a1d"
fi

# This simulates a pre-packaged Base64 string payload generated from encryption.ts
MOCK_ENCRYPTED_PAYLOAD="eyJlcGhlbWVyYWxQdWJLZXkiOiIwM2U5YSIsIml2IjoiYWYzZSIsImF1dGhUYWciOiI5OWJiIiwiY2lwaGVydGV4dCI6IjEyM2EifQ=="

# 2. Adjusted port to 3000 (assuming Go Fiber standard dev port to avoid LND's 8080 clash)
API_URL="http://localhost:3000/api/vaults"

echo "📡 Transmitting registration payload to local Fiber API endpoint ($API_URL)..."
echo "🔑 Target Beneficiary Pubkey: $TARGET_PUBKEY"

RESPONSE=$(curl -s -X POST "$API_URL" \
  -H "Content-Type: application/json" \
  -d "{
    \"alias\": \"CLI-Automated-Vault\",
    \"beneficiary_pubkey\": \"$TARGET_PUBKEY\",
    \"encrypted_payload\": \"$MOCK_ENCRYPTED_PAYLOAD\",
    \"check_in_interval_seconds\": 15
  }")

echo -e "\n📦 API Response Server Output:"
# Format with jq if available, otherwise raw print
if command -v jq &> /dev/null; then
    echo "$RESPONSE" | jq '.'
else
    echo "$RESPONSE"
fi