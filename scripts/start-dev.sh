#!/bin/bash
set -euo pipefail

# Resolve repository root (script lives in ./scripts)
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# Establish the structural codebase absolute root directory anchor
ROOT_DIR="/workspaces/Continuum"

echo "⛓️  Checking underlying infrastructure health status..."

# 1. Ensure Docker microservices are up and running before starting apps
if command -v docker &> /dev/null; then
    if ! docker ps | grep -q "continuum-bitcoind"; then
        echo "🐳 Infrastructure stack down. Spinning up environment cluster..."
        docker compose -f "$ROOT_DIR/infrastructure/docker/docker-compose.yml" up -d
        sleep 2
    fi
else
    echo "❌ Error: Docker daemon dependency missing. Please start Docker first."
    exit 1
fi

# 2. Avoid port collisions from a previously running API or web server
for port in 8080 3000; do
    if lsof -i ":$port" >/dev/null 2>&1; then
        echo "⚠️  Port $port is already in use. Stopping the existing listener to keep the dev stack healthy..."
        lsof -ti ":$port" | xargs -r kill -9
    fi
done

# 3. Clear stale Next.js development processes from earlier runs
if pgrep -af "[/]workspaces/Continuum/node_modules/.bin/next dev" >/dev/null 2>&1; then
    echo "🧹 Clearing stale Next.js development processes..."
    pkill -f "[/]workspaces/Continuum/node_modules/.bin/next dev" || true
fi

# 2. Source the central environment variables safely into this execution runtime shell
if [ -f "$ROOT_DIR/.env" ]; then
    echo "📝 Exporting central configuration environment flags..."
    export $(grep -v '^#' "$ROOT_DIR/.env" | xargs)
fi

echo "🚀 Launching Continuum Backend Microservice Instance (Go/Fiber)..."
cd "$ROOT_DIR/apps/api"
# Running with air (if installed) or standard go run
if command -v air &> /dev/null; then
    air &
else
    go run cmd/server/main.go &
fi
API_PID=$!

echo "⚛️  Starting Web Interface Server Engine (Next.js/Vite)..."
cd "$ROOT_DIR/apps/web"
PORT=3000 npm run dev -- --port 3000 &
WEB_PID=$!

# Ensure graceful containment breakdown on termination signals
trap "echo -e '\n🛑 Shutting down microservices gracefully...'; kill $API_PID $WEB_PID 2>/dev/null; exit" INT TERM EXIT

echo "=============================================================================="
echo "🎉 Development environment running. Press Ctrl+C to terminate all services safely."
echo "=============================================================================="

wait
