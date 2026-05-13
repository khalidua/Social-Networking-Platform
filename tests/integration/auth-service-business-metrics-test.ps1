param(
    [string]$BaseUrl = "http://localhost:8081"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Invoke-Endpoint {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Uri
    )

    try {
        $response = Invoke-WebRequest -Uri $Uri -Method Get
        return [pscustomobject]@{
            StatusCode = [int]$response.StatusCode
            Body       = [string]$response.Content
        }
    }
    catch [System.Net.WebException] {
        if ($null -eq $_.Exception.Response) {
            throw
        }

        $httpResponse = [System.Net.HttpWebResponse]$_.Exception.Response
        $reader = New-Object System.IO.StreamReader($httpResponse.GetResponseStream())
        $body = $reader.ReadToEnd()
        $reader.Dispose()

        return [pscustomobject]@{
            StatusCode = [int]$httpResponse.StatusCode
            Body       = [string]$body
        }
    }
}

function Get-MetricValue {
    param(
        [Parameter(Mandatory = $true)]
        [string]$MetricsBody,
        [Parameter(Mandatory = $true)]
        [string]$MetricName,
        [string]$Labels = ""
    )

    if ([string]::IsNullOrWhiteSpace($Labels)) {
        $pattern = "(?m)^" + [regex]::Escape($MetricName) + "\s+([0-9]+(?:\.[0-9]+)?)$"
    }
    else {
        $pattern = "(?m)^" + [regex]::Escape($MetricName) + "\{" + [regex]::Escape($Labels) + "\}\s+([0-9]+(?:\.[0-9]+)?)$"
    }

    $match = [regex]::Match($MetricsBody, $pattern)
    if (-not $match.Success) {
        return [double]0
    }

    return [double]$match.Groups[1].Value
}

Write-Host "Checking auth-service metrics endpoint at $BaseUrl/metrics ..."
$metricsBeforeResponse = Invoke-Endpoint -Uri "$BaseUrl/metrics"

if ($metricsBeforeResponse.StatusCode -eq 404) {
    throw "/metrics returned 404. Restart or rebuild auth-service so the promhttp route is loaded."
}

if ($metricsBeforeResponse.StatusCode -ne 200) {
    throw "/metrics returned unexpected status code $($metricsBeforeResponse.StatusCode)."
}

$durationCountBefore = Get-MetricValue `
    -MetricsBody $metricsBeforeResponse.Body `
    -MetricName "auth_service_business_operation_duration_seconds_count" `
    -Labels 'operation="authenticate_user"'

$failureCountBefore = Get-MetricValue `
    -MetricsBody $metricsBeforeResponse.Body `
    -MetricName "auth_service_business_operation_total" `
    -Labels 'status="failure"'

$successCountBefore = Get-MetricValue `
    -MetricsBody $metricsBeforeResponse.Body `
    -MetricName "auth_service_business_operation_total" `
    -Labels 'status="success"'

Write-Host "Triggering AuthenticateUser failure path via /api/v1/auth/callback ..."
$callbackResponse = Invoke-Endpoint -Uri "$BaseUrl/api/v1/auth/callback"

if ($callbackResponse.StatusCode -ne 400) {
    throw "Expected /api/v1/auth/callback to return 400 for missing code/state, got $($callbackResponse.StatusCode). Body: $($callbackResponse.Body)"
}

Write-Host "Re-reading metrics ..."
$metricsAfterResponse = Invoke-Endpoint -Uri "$BaseUrl/metrics"

if ($metricsAfterResponse.StatusCode -ne 200) {
    throw "Second /metrics request returned unexpected status code $($metricsAfterResponse.StatusCode)."
}

$durationCountAfter = Get-MetricValue `
    -MetricsBody $metricsAfterResponse.Body `
    -MetricName "auth_service_business_operation_duration_seconds_count" `
    -Labels 'operation="authenticate_user"'

$failureCountAfter = Get-MetricValue `
    -MetricsBody $metricsAfterResponse.Body `
    -MetricName "auth_service_business_operation_total" `
    -Labels 'status="failure"'

$successCountAfter = Get-MetricValue `
    -MetricsBody $metricsAfterResponse.Body `
    -MetricName "auth_service_business_operation_total" `
    -Labels 'status="success"'

if ($durationCountAfter -le $durationCountBefore) {
    throw "Expected histogram count to increase. Before=$durationCountBefore After=$durationCountAfter"
}

if ($failureCountAfter -le $failureCountBefore) {
    throw "Expected failure counter to increase. Before=$failureCountBefore After=$failureCountAfter"
}

Write-Host ""
Write-Host "PASS: auth-service business metrics are updating inside service logic."
Write-Host "Histogram count: $durationCountBefore -> $durationCountAfter"
Write-Host "Failure count : $failureCountBefore -> $failureCountAfter"
Write-Host "Success count : $successCountBefore -> $successCountAfter"
