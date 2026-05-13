param(
    [string]$ComposeFile = (Join-Path $PSScriptRoot "..\compose\compose.yml"),
    [switch]$Build
)

$ErrorActionPreference = "Stop"
$argsList = @("compose", "-f", $ComposeFile, "up", "-d")
if ($Build) {
    $argsList += "--build"
}
& docker @argsList
if ($LASTEXITCODE -ne 0) {
    throw "docker compose up failed"
}
