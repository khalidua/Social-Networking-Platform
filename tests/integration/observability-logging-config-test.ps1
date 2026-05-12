$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..\..")

function Assert-Contains {
    param(
        [string] $Path,
        [string] $Expected
    )

    $content = Get-Content -Raw -Path $Path
    if (-not $content.Contains($Expected)) {
        throw "Expected '$Path' to contain '$Expected'"
    }
}

$compose = Join-Path $root "deploy\compose\compose.yml"
$loki = Join-Path $root "deploy\logging\loki\loki.yml"
$promtail = Join-Path $root "deploy\logging\promtail\promtail.yml"
$datasources = Join-Path $root "deploy\monitoring\grafana\provisioning\datasources\prometheus.yml"
$logsDashboard = Join-Path $root "deploy\monitoring\grafana\dashboards\social-networking-platform-logs.json"

foreach ($path in @($compose, $loki, $promtail, $datasources, $logsDashboard)) {
    if (-not (Test-Path $path)) {
        throw "Required logging config file is missing: $path"
    }
}

Assert-Contains $compose "loki:"
Assert-Contains $compose "promtail:"
Assert-Contains $compose "3100:3100"
Assert-Contains $compose "9080:9080"
Assert-Contains $loki "auth_enabled: false"
Assert-Contains $promtail "docker_sd_configs:"
Assert-Contains $promtail "http://loki:3100/loki/api/v1/push"
Assert-Contains $promtail "route_group:"
Assert-Contains $promtail "status_group:"
Assert-Contains $promtail "request_id:"
Assert-Contains $datasources "uid: Loki"

Get-Content -Raw -Path $logsDashboard | ConvertFrom-Json | Out-Null

if (Get-Command docker -ErrorAction SilentlyContinue) {
    Push-Location $root
    try {
        docker compose -f deploy\compose\compose.yml config --quiet
    }
    finally {
        Pop-Location
    }
}

Write-Output "centralized logging config ok"
