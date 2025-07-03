#!/bin/bash

# Script to generate Go code from protocol buffer definitions

set -e

# Create output directory
mkdir -p proto/gen

# Install protoc plugins if not present
echo "Installing protoc plugins..."
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

# Generate Go code from proto files
echo "Generating Go code from protocol buffers..."

protoc --go_out=proto/gen --go_opt=paths=source_relative \
       --go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
       --grpc-gateway_out=proto/gen --grpc-gateway_opt=paths=source_relative \
       proto/functions.proto

echo "✅ Protocol buffer code generation completed!"
echo "Generated files:"
find proto/gen -name "*.go" -type f 