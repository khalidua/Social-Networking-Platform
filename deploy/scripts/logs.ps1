param(
    [string]$ComposeFile = (Join-Path $PSScriptRoot "..\compose\compose.yml"),
    [string]$Service = ""
)

$ErrorActionPreference = "Stop"
if ([string]::IsNullOrWhiteSpace($Service)) {
    & docker compose -f $ComposeFile logs --tail 200
} else {
    & docker compose -f $ComposeFile logs --tail 200 $Service
}
if ($LASTEXITCODE -ne 0) {
    throw "docker compose logs failed"
}
