param(
    [string]$ComposeFile = (Join-Path $PSScriptRoot "..\compose\compose.yml")
)

$ErrorActionPreference = "Stop"
& docker compose -f $ComposeFile down
if ($LASTEXITCODE -ne 0) {
    throw "docker compose down failed"
}
