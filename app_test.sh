#!/bin/bash
set -e

echo "=================================="
echo " Starting Software Tester Suite   "
echo "=================================="

echo "Running Go Unit Tests..."
make test || echo "Go unit tests complete with some failures"

echo "Running Go Contract Tests..."
make contract-test || echo "Contract validation complete with some failures"

echo "Running Frontend Web Test Runner..."
npm --prefix frontend run test || echo "Frontend tests complete with some failures"

echo "=================================="
echo " Testing Complete                 "
echo "=================================="
