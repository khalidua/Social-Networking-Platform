param(
    [string]$ComposeFile = (Join-Path $PSScriptRoot "..\compose\compose.yml"),
    [switch]$ConfirmVolumeDelete
)

$ErrorActionPreference = "Stop"
if (-not $ConfirmVolumeDelete) {
    throw "Refusing to delete Compose volumes without -ConfirmVolumeDelete"
}

& docker compose -f $ComposeFile down -v
if ($LASTEXITCODE -ne 0) {
    throw "docker compose down -v failed"
}
