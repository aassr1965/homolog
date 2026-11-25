<#
.SYNOPSIS
    Script de pre-validation qualite et conformite SANAD PHARMA (Go + Frontend JS + GAMP 5).
.PARAMETER Full
    Execute egalement govulncheck et golangci-lint si disponibles.
#>
param(
    [switch]$Full
)

$ErrorActionPreference = "Stop"

function Write-Step($msg) {
    Write-Host "`n==> $msg" -ForegroundColor Cyan
}

# 1. Formatage Go strict (gofmt)
Write-Step "1/5 Formatage Go (gofmt -w .)"
gofmt -w .
if ($LASTEXITCODE -ne 0) { throw "gofmt a echoue." }

# 2. Analyse statique Go (go vet)
Write-Step "2/5 Analyse statique Go (go vet ./...)"
go vet ./...
if ($LASTEXITCODE -ne 0) { throw "go vet a detecte des anomalies." }

# 3. Verification syntaxique JavaScript (node --check)
Write-Step "3/5 Verification syntaxique JavaScript (web/js/*.js)"
$nodeCmd = Get-Command node -ErrorAction SilentlyContinue
if ($nodeCmd) {
    $jsFiles = Get-ChildItem -Path "web/js" -Filter "*.js"
    foreach ($file in $jsFiles) {
        & node --check $file.FullName
        if ($LASTEXITCODE -ne 0) { throw "Erreur syntaxique JS dans $($file.Name)" }
    }
    Write-Host "    $($jsFiles.Count) fichiers JS verifies sans erreur." -ForegroundColor Green
} else {
    Write-Warning "node.exe introuvable : verification syntaxique JS ignoree."
}

# 4. Suite de tests Go
Write-Step "4/5 Execution des tests Go (go test ./...)"
$gccCmd = Get-Command gcc -ErrorAction SilentlyContinue
if ($gccCmd -and ($env:CGO_ENABLED -eq "1")) {
    go test -race ./...
} else {
    go test ./...
}
if ($LASTEXITCODE -ne 0) { throw "Des tests Go ont echoue." }

# 5. Compilation de controle (incluant //go:embed web/)
Write-Step "5/5 Verification de compilation (go build)"
$tempBin = Join-Path $env:TEMP "sanad_verify_server.exe"
go build -o $tempBin ./cmd/server
if ($LASTEXITCODE -ne 0) {
    if (Test-Path $tempBin) { Remove-Item $tempBin -Force -ErrorAction SilentlyContinue }
    throw "La compilation du serveur Go a echoue."
}
if (Test-Path $tempBin) { Remove-Item $tempBin -Force -ErrorAction SilentlyContinue }
Write-Host "    Compilation et embed web/ valides." -ForegroundColor Green

# Optionnel : Audit securite & dependances (-Full)
if ($Full) {
    Write-Step "Bonus (-Full) : Audit des vulnerabilites (govulncheck) & Linter"
    if (Get-Command govulncheck -ErrorAction SilentlyContinue) {
        govulncheck ./...
        if ($LASTEXITCODE -ne 0) { throw "govulncheck a signale des vulnerabilites." }
    }
    if (Get-Command golangci-lint -ErrorAction SilentlyContinue) {
        golangci-lint run
        if ($LASTEXITCODE -ne 0) { throw "golangci-lint a signale des avertissements." }
    }
}

Write-Host "`n[OK] Toutes les verifications SANAD PHARMA ont reussi !" -ForegroundColor Green
