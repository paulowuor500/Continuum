#!/bin/bash
set -e

# Define a short helper function to execute RPC interactions cleanly
btc-cli() {
  docker exec continuum-bitcoind bitcoin-cli -regtest -rpcuser=continuum -rpcpassword=finality "$@"
}

echo "⏳ Probing Bitcoin Core RPC layer until interface stabilizes..."
# Poll the node until it stops returning errors, confirming it is ready to accept commands
until btc-cli getblockchaininfo > /dev/null 2>&1; do
  echo "   [Waiting for bitcoind to initialize internal databases...]"
  sleep 1
done

echo "🪙 Setting up developer regtest mining wallet..."
# Check if the wallet is already loaded; if not, attempt to create it
if ! btc-cli listwallets | grep -q "continuum_miner"; then
  echo "🤖 Wallet 'continuum_miner' not detected. Creating fresh instance..."
  btc-cli createwallet continuum_miner || btc-cli loadwallet continuum_miner
fi

echo "⛏️ Mining 101 initial blocks to mature block rewards..."
# Target the established mining wallet context explicitly
MINING_ADDR=$(btc-cli -rpcwallet=continuum_miner getnewaddress)
btc-cli -rpcwallet=continuum_miner generatetoaddress 101 "$MINING_ADDR"

echo "=============================================================================="
echo "✅ SUCCESS: Chain state fully synchronized!"
echo "📊 Current Block Count: $(btc-cli getblockcount)"
echo "💰 Miner Balance:       $(btc-cli -rpcwallet=continuum_miner getbalance) BTC"
echo "=============================================================================="

echo "🏃‍♂️ Transitioning directly to wallet allocation routines..."
# Make sure wallets.sh is executable and preserve execution path integrity
chmod +x ./wallets.sh
./wallets.sh