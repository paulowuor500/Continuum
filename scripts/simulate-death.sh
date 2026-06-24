#!/bin/bash

# Exit on error
set -e

VAULT_ID=${1:-"demo-vault-uuid-placeholder"}
API_PORT=${2:-"3000"} # Defaulting to 3000 to avoid LND port 8080 clash

echo "☠️  Initiating Liveness Failure Simulation for Vault: $VAULT_ID"
echo "------------------------------------------------------------------"

# 1. Advance the actual Blockchain layer to simulate time passing
if command -v docker &> /dev/null && docker ps | grep -q continuum-bitcoind; then
    echo "⛏️  Mining 30 quick blocks on Regtest to mature time/lock parameters..."
    # Mining blocks forces block-height-based timelocks to expire immediately
    docker exec continuum-bitcoind bitcoin-cli -regtest -rpcuser=continuum -rpcpassword=finality generate 30 > /dev/null
    echo "🔗 Regtest chain advanced cleanly."
else
    echo "⚠️  Warning: continuum-bitcoind container not found. Skipping physical chain warp."
fi

# 2. Notify the Fiber Application Engine
API_URL="http://localhost:$API_PORT/api/vaults/$VAULT_ID/warp"
echo "📡 Sending displacement network trigger to application layer ($API_URL)..."

RESPONSE=$(curl -s -X POST "$API_URL")

# Flexible validation (checks for 'SUCCESS' or typical JSON success fields)
if echo "$RESPONSE" | grep -qiE "(SUCCESS|true|status)"; then
    echo "✅ Time warp simulation acknowledged by backend runtime!"
    echo "📦 Server response: $RESPONSE"
    echo "⏳ Your Go background monitor routine will flag this vault DORMANT on its next cron interval."
else
    echo "❌ Execution failure. Ensure your Fiber engine is active on port $API_PORT."
    echo "📦 Server output: $RESPONSE"
fi