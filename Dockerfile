# ----------- Build Go orchestrator -----------
FROM golang:1.21 AS go-builder
WORKDIR /app/go-orchestrator
COPY go-orchestrator/go.mod go-orchestrator/go.sum ./
RUN go mod download
COPY go-orchestrator/ .
RUN go build -o /app/bin/go-orchestrator main.go

# ----------- Build Node renderer/client -----------
FROM node:18-slim AS node-builder
WORKDIR /app

# Copy package manager and workspace files first to leverage caching
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
COPY tsconfig.json ./

# Copy workspace package.json files
COPY user-app/package.json ./user-app/
COPY node-renderer/package.json ./node-renderer/

# Install all dependencies for the workspace
RUN npm install -g pnpm && pnpm install --frozen-lockfile --recursive

# Copy the rest of the source code for node-related projects
COPY user-app/ ./user-app/
COPY node-renderer/ ./node-renderer/

# Build the client assets and server functions
RUN pnpm --filter user-app run generate:types && \
    pnpm --filter node-renderer generate:import-map && \
    pnpm --filter node-renderer build:client && \
    pnpm --filter node-renderer build:server && \
    pnpm --filter user-app build:server

# ----------- Build Go plugins -----------
FROM golang:1.21 AS go-plugins-builder
WORKDIR /app
COPY user-app ./user-app
COPY go-orchestrator/build-plugins ./go-orchestrator/build-plugins
WORKDIR /app/go-orchestrator
RUN go run build-plugins/main.go

# ----------- Final minimal image -----------
FROM node:18-slim
WORKDIR /app
# Install netcat for port checks
RUN apt-get update && apt-get install -y netcat-openbsd && rm -rf /var/lib/apt/lists/*
# Copy Go binary
COPY --from=go-builder /app/bin/go-orchestrator ./bin/go-orchestrator
# Copy built Node renderer/client and user-app
COPY --from=node-builder /app/user-app ./user-app
COPY --from=node-builder /app/node-renderer ./node-renderer
COPY --from=node-builder /app/node_modules ./node_modules
COPY --from=node-builder /app/package.json ./package.json
COPY --from=node-builder /app/pnpm-lock.yaml ./pnpm-lock.yaml
# Copy Go plugins
COPY --from=go-plugins-builder /app/user-app/server/go/*.so ./user-app/server/go/
# Copy compiled TS server functions
COPY --from=node-builder /app/user-app/dist/ts ./user-app/dist/ts
# Copy TS server functions (for dev mode if needed, and for ts-node to find original sources if it needs them)
COPY user-app/server/ts ./user-app/server/ts
# Add entrypoint script
COPY docker-entrypoint.sh ./docker-entrypoint.sh
RUN chmod +x ./docker-entrypoint.sh
RUN npm install -g pnpm
# Expose Go and Node ports
EXPOSE 8080 3001
# Use non-root user for security
RUN useradd -m appuser && chown -R appuser /app
USER appuser
# Debug: list server function files
RUN ls -lR ./user-app/server/go || true
RUN ls -lR ./user-app/server/ts || true
# Debug: check ts-node version and list TS server functions
RUN npx ts-node --version || true
RUN ls -lR ./user-app/server/ts || true
# Debug: list compiled TS server function files
RUN ls -lR ./user-app/dist/ts || true
# Start both servers
ENTRYPOINT ["./docker-entrypoint.sh"] 