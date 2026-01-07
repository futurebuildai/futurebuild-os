#!/bin/bash

# FutureBuild Directory Generator
# This script establishes the Standard Go Layout and Lit Frontend structure.

echo "🏗️  Scaffolding FutureBuild Project Structure..."

# 1. Root and Go Backend Structure
mkdir -p cmd/api
mkdir -p internal/data
mkdir -p internal/server
mkdir -p pkg/types

# 2. Frontend Structure
mkdir -p frontend/src/components
mkdir -p frontend/src/styles
mkdir -p frontend/src/types
mkdir -p frontend/src/artifacts
mkdir -p frontend/public

# 3. Create Placeholder files if they don't exist
touch cmd/api/main.go
touch internal/server/server.go

echo "✅ Structure Generated Successfully!"
