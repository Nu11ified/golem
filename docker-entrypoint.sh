#!/bin/bash
set -e

# Start Go orchestrator
./bin/go-orchestrator &
GO_PID=$!

# Start Node renderer
pnpm --filter node-renderer start &
NODE_PID=$!

# Wait for both processes
wait $GO_PID
wait $NODE_PID 