#!/bin/bash

# Exit on error
set -e

VAULT_ID=${1:-""}
API_PORT=${2:-"8080"}

resolve_vault_id() {
    if [[ -n "$VAULT_ID" ]] && [[ "$VAULT_ID" =~ ^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$ ]]; then
        echo "$VAULT_ID"
        return 0
    fi

    echo "⚠️  No valid vault UUID provided. Looking up the first available vault from the API..." >&2
    if command -v jq >/dev/null 2>&1; then
        jq -r '.vaults[0].id // empty' <(curl -s "http://localhost:$API_PORT/api/vaults") 2>/dev/null
    else
        python3 - <<'PY' "$API_PORT"
import json, sys, urllib.request
url = f"http://localhost:{sys.argv[1]}/api/vaults"
try:
    with urllib.request.urlopen(url, timeout=5) as resp:
        data = json.load(resp)
    vaults = data.get("vaults", [])
    if vaults:
        print(vaults[0].get("id", ""))
except Exception:
    pass
PY
    fi
}

TARGET_VAULT_ID=$(resolve_vault_id)
if [[ -z "$TARGET_VAULT_ID" ]]; then
    echo "❌ No vault available to simulate. Create a vault first and rerun the script." >&2
    exit 1
fi

echo "☠️  Initiating Liveness Failure Simulation for Vault: $TARGET_VAULT_ID"
echo "------------------------------------------------------------------"

# 1. Advance the actual Blockchain layer to simulate time passing
if command -v docker &> /dev/null && docker ps | grep -q continuum-bitcoind; then
    echo "⛏️  Ensuring a regtest wallet is available before mining..."
    docker exec continuum-bitcoind bitcoin-cli -regtest -rpcuser=continuum -rpcpassword=finality createwallet "continuum-wallet" >/dev/null 2>&1 || true
    docker exec continuum-bitcoind bitcoin-cli -regtest -rpcuser=continuum -rpcpassword=finality loadwallet "continuum-wallet" >/dev/null 2>&1 || true

    echo "⛏️  Mining 30 quick blocks on Regtest to mature time/lock parameters..."
    # Mining blocks forces block-height-based timelocks to expire immediately
    docker exec continuum-bitcoind bitcoin-cli -regtest -rpcuser=continuum -rpcpassword=finality -generate 30 > /dev/null
    echo "🔗 Regtest chain advanced cleanly."
else
    echo "⚠️  Warning: continuum-bitcoind container not found. Skipping physical chain warp."
fi

# 2. Notify the Fiber Application Engine
API_URL="http://localhost:$API_PORT/api/vaults/$TARGET_VAULT_ID/warp"
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