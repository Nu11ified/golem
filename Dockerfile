# ----------- Build Go orchestrator -----------
FROM golang:1.21-alpine AS go-builder
WORKDIR /app/go-orchestrator
COPY go-orchestrator/go.mod go-orchestrator/go.sum ./
RUN go mod download
COPY go-orchestrator/ .
RUN go build -o /app/bin/go-orchestrator main.go

# ----------- Build Node renderer/client -----------
FROM node:18-alpine AS node-builder
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
COPY user-app/package.json user-app/
COPY node-renderer/package.json node-renderer/
RUN npm install -g pnpm && pnpm install --frozen-lockfile --recursive
COPY user-app/ user-app/
COPY node-renderer/ node-renderer/
RUN pnpm --filter user-app run generate:types && \
    pnpm --filter node-renderer generate:import-map && \
    pnpm --filter node-renderer build:client

# ----------- Final minimal image -----------
FROM node:18-slim
WORKDIR /app
# Copy Go binary
COPY --from=go-builder /app/bin/go-orchestrator ./bin/go-orchestrator
# Copy built Node renderer/client and user-app
COPY --from=node-builder /app/user-app ./user-app
COPY --from=node-builder /app/node-renderer ./node-renderer
COPY --from=node-builder /app/node_modules ./node_modules
COPY --from=node-builder /app/package.json ./package.json
COPY --from=node-builder /app/pnpm-lock.yaml ./pnpm-lock.yaml
# Add entrypoint script
COPY docker-entrypoint.sh ./docker-entrypoint.sh
RUN chmod +x ./docker-entrypoint.sh
# Expose Go and Node ports
EXPOSE 8080 3001
# Use non-root user for security
RUN useradd -m appuser && chown -R appuser /app
USER appuser
# Start both servers
ENTRYPOINT ["./docker-entrypoint.sh"] 