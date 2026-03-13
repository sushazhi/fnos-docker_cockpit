#!/bin/bash

echo "=== Checking Frontend Dependencies ==="
cd app/ui
echo "Checking for outdated packages..."
npm outdated || true

echo ""
echo "Checking for security vulnerabilities..."
npm audit

echo ""
echo "=== Checking Backend Dependencies ==="
cd ../server
echo "Checking for outdated Go modules..."
go list -u -m all | grep '\['

echo ""
echo "Checking for known vulnerabilities..."
go list -json -m all | nancy sleuth || echo "Install nancy: go install github.com/sonatype-nexus-community/nancy@latest"

echo ""
echo "=== Dependency Check Complete ==="
