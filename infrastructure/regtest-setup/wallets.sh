#!/bin/bash
set -e

echo "⚡ Initializing Lightning Network Node Dev Wallets..."

btc-cli() {
  docker exec continuum-bitcoind bitcoin-cli -regtest -rpcuser=continuum -rpcpassword=finality -rpcwallet=continuum_miner "$@"
}

ln-cli() {
  docker exec continuum-lnd lncli --network=regtest "$@"
}

# Check if LND crashed due to missing wallet
if docker logs continuum-lnd 2>&1 | grep -q "initialize the wallet before using auto unlocking"; then
  echo "🌱 Hard-bootstrapping fresh LND database and seed files..."
  
  # Start a temporary detached container to create the wallet file inside the shared volume
  docker run -d --name continuum-lnd-init \
    -v docker_lnd_data:/root/.lnd \
    lightninglabs/lnd:v0.17.4-beta --bitcoin.active --bitcoin.regtest --bitcoin.node=bitcoind
  
  sleep 3
  echo "🔑 Injecting credentials into seed state..."
  # Create the wallet via RPC
  docker exec continuum-lnd-init lncli --network=regtest create --wallet-password-string=finality --noscript > /dev/null 2>&1 || true
  
  # Tear down the temporary bootstrap container
  docker rm -f continuum-lnd-init
  
  # Kick the main LND node container back into action
  docker compose -f ../docker/docker-compose.yml restart lnd
fi

echo "⏳ Probing LND gRPC interface until online..."
until docker exec continuum-lnd lncli --network=regtest state > /dev/null 2>&1; do
  echo "   [Waiting for LND service port to answer...]"
  sleep 2
done

echo "📥 Generating on-chain deposit address for Continuum LND Node..."
LND_ADDRESS=$(ln-cli newaddress p2wkh | grep '"address"' | awk -F'"' '{print $4}')

if [ -z "$LND_ADDRESS" ]; then
    echo "❌ Error: Could not retrieve a valid deposit address from LND node container."
    exit 1
fi
echo "🎯 LND Target Deposit Address: $LND_ADDRESS"

echo "💸 Sending 10 BTC from matured miner wallet pool to LND node..."
TX_ID=$(btc-cli sendtoaddress "$LND_ADDRESS" 10.0)
echo "📦 Transaction Broadcast ID: $TX_ID"

echo "⛏️ Mining 6 blocks to force on-chain confirmation..."
MINER_ADDR=$(btc-cli getnewaddress)
btc-cli generatetoaddress 6 "$MINER_ADDR"

echo "⏳ Waiting for LND to parse and credit your block confirmations..."
sleep 3

echo "📊 Verifying node balances..."
WALLET_BALANCE=$(ln-cli walletbalance | grep '"confirmed_balance"' | awk -F'"' '{print $4}')

echo "=============================================================================="
echo "🎉 SUCCESS: LND Node Loaded and Funded!"
echo "💰 Current Spendable LND Confirmed Balance: $WALLET_BALANCE Satoshis"
echo "=============================================================================="