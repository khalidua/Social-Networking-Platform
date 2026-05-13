param(
    [string]$GatewayBaseUrl = "http://localhost:8080"
)

$ErrorActionPreference = "Stop"
$targets = @(
    @{ Name = "api-gateway"; Url = "$GatewayBaseUrl/health" },
    @{ Name = "auth-service"; Url = "http://localhost:8081/health" },
    @{ Name = "users-service"; Url = "http://localhost:8082/health" },
    @{ Name = "posts-service"; Url = "http://localhost:8083/health" },
    @{ Name = "feed-service"; Url = "http://localhost:8084/health" },
    @{ Name = "notification-service"; Url = "http://localhost:8085/health" }
)

foreach ($target in $targets) {
    $response = Invoke-WebRequest -Uri $target.Url -UseBasicParsing
    if ([int]$response.StatusCode -ne 200) {
        throw "$($target.Name) health returned $($response.StatusCode)"
    }
    Write-Host "PASS $($target.Name) $($target.Url)"
}
