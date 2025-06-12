# ----------- Build Go orchestrator -----------
FROM golang:1.21-alpine AS go-builder
RUN apk add --no-cache build-base
WORKDIR /app/go-orchestrator
COPY go-orchestrator/go.mod go-orchestrator/go.sum ./
RUN go mod download
COPY go-orchestrator/ .
# Copy user-app server functions for plugin build
COPY user-app/server/go /app/user-app/server/go
COPY user-app/server/ts /app/user-app/server/ts
RUN go build -o /app/bin/go-orchestrator main.go
# Build Go plugins
RUN go run build-plugins/main.go

# ----------- Build Node renderer/client -----------
FROM node:18-alpine AS node-builder
WORKDIR /app

# Copy package manager and workspace files first to leverage caching
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./

# Copy workspace package.json files
COPY user-app/package.json ./user-app/
COPY node-renderer/package.json ./node-renderer/

# Install all dependencies for the workspace
RUN npm install -g pnpm && pnpm install --frozen-lockfile --recursive

# Copy the rest of the source code for node-related projects
COPY user-app/ ./user-app/
COPY node-renderer/ ./node-renderer/

# Build the client assets
RUN pnpm --filter user-app run generate:types && \
    pnpm --filter node-renderer generate:import-map && \
    pnpm --filter node-renderer build:client

# ----------- Final minimal image -----------
FROM node:18-slim
WORKDIR /app
# Copy Go binary
COPY --from=go-builder /app/bin/go-orchestrator ./bin/go-orchestrator
# Copy built Go plugins
COPY --from=go-builder /app/user-app/server/go ./user-app/server/go
# Copy built Node renderer/client and user-app
COPY --from=node-builder /app/user-app ./user-app
COPY --from=node-builder /app/node-renderer ./node-renderer
COPY --from=node-builder /app/node_modules ./node_modules
COPY --from=node-builder /app/package.json ./package.json
COPY --from=node-builder /app/pnpm-lock.yaml ./pnpm-lock.yaml
# Copy TypeScript server functions
COPY --from=node-builder /app/user-app/server/ts ./user-app/server/ts
# Copy Node renderer function runner
COPY --from=node-builder /app/node-renderer/ts-function-runner.js ./node-renderer/ts-function-runner.js
# Copy Node renderer dist (for client.js)
COPY --from=node-builder /app/node-renderer/dist ./node-renderer/dist
# Add entrypoint script
COPY docker-entrypoint.sh ./docker-entrypoint.sh
RUN chmod +x ./docker-entrypoint.sh
# Install pnpm globally for entrypoint
RUN npm install -g pnpm
# Expose Go and Node ports
EXPOSE 8080 3001
# Use non-root user for security
RUN useradd -m appuser && chown -R appuser /app
USER appuser
# Start both servers
ENTRYPOINT ["./docker-entrypoint.sh"] 