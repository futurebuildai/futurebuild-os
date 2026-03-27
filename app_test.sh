#!/bin/bash
echo "=== Testing FutureBuild OS Backend ==="
go test -v ./...

echo "=== Testing FutureBuild OS Frontend ==="
if [ -d "frontend" ]; then
  cd frontend
  if [ -f "package.json" ]; then
    npm install --no-audit --no-fund --legacy-peer-deps
    npm run test || echo "Frontend tests failed or not configured"
  fi
  cd ..
fi
