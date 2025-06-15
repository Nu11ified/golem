#!/bin/bash
set -e

# Start Node renderer first
pnpm --filter node-renderer start &
NODE_PID=$!

# Wait for Node renderer to be ready
NODE_PORT="${NODE_PORT:-3001}"
echo "Waiting for Node renderer to be ready on port $NODE_PORT..."
while ! nc -z localhost "$NODE_PORT"; do
  sleep 0.5
done
echo "Node renderer is up!"

# Start Go orchestrator
./bin/go-orchestrator &
GO_PID=$!

# Wait for both processes
wait $GO_PID
wait $NODE_PID 