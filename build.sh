#!/bin/zsh

# Build the project
echo "Building the project..."
go build -o goloadtest ./cmd/main.go
echo "Project built successfully!"
