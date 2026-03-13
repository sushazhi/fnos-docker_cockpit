# Frontend Dependencies Check
Write-Host "=== Checking Frontend Dependencies ===" -ForegroundColor Cyan
Set-Location app/ui

Write-Host "Checking for outdated packages..." -ForegroundColor Yellow
npm outdated

Write-Host ""
Write-Host "Checking for security vulnerabilities..." -ForegroundColor Yellow
npm audit

# Backend Dependencies Check
Write-Host ""
Write-Host "=== Checking Backend Dependencies ===" -ForegroundColor Cyan
Set-Location ../server

Write-Host "Checking for outdated Go modules..." -ForegroundColor Yellow
go list -u -m all | Select-String '\['

Write-Host ""
Write-Host "Checking for known vulnerabilities..." -ForegroundColor Yellow
Write-Host "Install nancy: go install github.com/sonatype-nexus-community/nancy@latest" -ForegroundColor Gray
go list -json -m all | nancy sleuth 2>$null

Write-Host ""
Write-Host "=== Dependency Check Complete ===" -ForegroundColor Green
Set-Location ../..
