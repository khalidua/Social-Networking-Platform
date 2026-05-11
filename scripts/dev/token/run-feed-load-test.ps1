param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$K6Script = "feed-service/tests/load/feed-load-test.js",
    [int]$Vus = 50,
    [string]$Duration = "30s",
    [switch]$SkipSessionSeed,
    [string]$SessionId = "session-1",
    [string]$UserId = "google:123",
    [string]$Email = "user@example.com",
    [int]$TtlSeconds = 3600,
    [string]$RedisContainer = "snp-redis"
)

$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..")).Path
$tokenGenerator = Join-Path $repoRoot "scripts/dev/token/generate-token.go"
$seedScript = Join-Path $repoRoot "scripts/dev/token/seed-session.ps1"
$k6ScriptPath = Join-Path $repoRoot $K6Script

if (-not (Test-Path $tokenGenerator)) {
    throw "Token generator not found at '$tokenGenerator'."
}
if (-not (Test-Path $k6ScriptPath)) {
    throw "k6 script not found at '$k6ScriptPath'."
}

Write-Host "Generating JWT token..."
$token = (go run $tokenGenerator | Select-Object -Last 1).Trim()
if (-not $token -or ($token -notmatch "^[^.]+\.[^.]+\.[^.]+$")) {
    throw "Generated token is empty or malformed."
}

if (-not $SkipSessionSeed) {
    Write-Host "Seeding session '$SessionId' in Redis..."
    & $seedScript `
        -SessionId $SessionId `
        -UserId $UserId `
        -Email $Email `
        -TtlSeconds $TtlSeconds `
        -RedisContainer $RedisContainer
}

Write-Host "Running k6 test..."
& k6 run $k6ScriptPath --vus $Vus --duration $Duration --env BASE_URL=$BaseUrl --env TOKEN=$token
