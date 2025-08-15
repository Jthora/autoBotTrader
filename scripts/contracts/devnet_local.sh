#!/usr/bin/env bash
set -euo pipefail

# Simple deterministic Starknet devnet launcher for local, zero-cost testing.
# Requires Docker. Provides free pre-funded accounts so no real financing needed.

IMAGE="shardlabs/starknet-devnet:latest"
CONTAINER_NAME="starknet-devnet-local"
PORT="5050"
SEED="20240815"
GAS_PRICE="1"  # Keep deterministic; adjust if image flag changes.

echo "[devnet] Starting local Starknet devnet (container: $CONTAINER_NAME, port: $PORT)"

if docker ps -a --format '{{.Names}}' | grep -Eq "^${CONTAINER_NAME}$"; then
  echo "[devnet] Removing existing container ${CONTAINER_NAME}" >&2
  docker rm -f "$CONTAINER_NAME" >/dev/null 2>&1 || true
fi

docker run -d \
  --name "$CONTAINER_NAME" \
  -p ${PORT}:5050 \
  "$IMAGE" \
  --seed ${SEED} \
  --gas-price ${GAS_PRICE} \
  --dump-on SIGINT \
  --accounts 10 \
  --lite-mode

echo "[devnet] Waiting for readiness..."
RETRIES=30
until curl -s "http://localhost:${PORT}/is_alive" >/dev/null 2>&1; do
  sleep 0.5
  RETRIES=$((RETRIES-1))
  if [ "$RETRIES" -le 0 ]; then
    echo "[devnet] Failed to start devnet" >&2
    docker logs "$CONTAINER_NAME" || true
    exit 1
  fi
done

echo "[devnet] Ready at http://localhost:${PORT}"
echo "[devnet] Sample accounts:" 
curl -s "http://localhost:${PORT}/predeployed_accounts" | head -n 40

echo "[devnet] To stop: docker rm -f ${CONTAINER_NAME}" 
