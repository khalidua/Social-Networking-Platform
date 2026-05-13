param(
    [string]$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..")).Path,
    [string]$ComposeFile = (Join-Path $RepoRoot "deploy\compose\compose.yml"),
    [string]$GatewayBaseUrl = "http://localhost:8080",
    [string]$JwtSecret = "change-me",
    [int]$ReadVus = 10,
    [string]$ReadDuration = "45s",
    [int]$WriteVus = 5,
    [string]$WriteDuration = "30s",
    [ValidateSet("auto", "local", "docker")]
    [string]$K6Runner = "auto",
    [string]$K6DockerImage = "grafana/k6:0.53.0",
    [switch]$StartStack,
    [switch]$SkipRead,
    [switch]$SkipWrite,
    [switch]$ValidateOnly
)

$ErrorActionPreference = "Stop"

$readScript = Join-Path $RepoRoot "tests\load\k6\gateway-read-load.js"
$writeScript = Join-Path $RepoRoot "tests\load\k6\social-write-stress.js"
$reportsDir = Join-Path $RepoRoot "tests\load\reports"

function Assert-True {
    param([bool]$Condition, [string]$Message)
    if (-not $Condition) {
        throw $Message
    }
}

function ConvertTo-Base64Url {
    param([byte[]]$Bytes)
    return [Convert]::ToBase64String($Bytes).TrimEnd("=").Replace("+", "-").Replace("/", "_")
}

function New-TestJwt {
    param([string]$UserID, [string]$Email, [string]$SessionID, [string]$Secret)
    $now = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
    $expires = [DateTimeOffset]::UtcNow.AddHours(1).ToUnixTimeSeconds()
    $headerJson = @{ alg = "HS256"; typ = "JWT" } | ConvertTo-Json -Compress
    $payloadJson = @{
        iss   = "auth-service"
        sub   = $UserID
        sid   = $SessionID
        email = $Email
        iat   = $now
        exp   = $expires
    } | ConvertTo-Json -Compress

    $header = ConvertTo-Base64Url ([Text.Encoding]::UTF8.GetBytes($headerJson))
    $payload = ConvertTo-Base64Url ([Text.Encoding]::UTF8.GetBytes($payloadJson))
    $signingInput = "$header.$payload"
    $hmac = [Security.Cryptography.HMACSHA256]::new([Text.Encoding]::UTF8.GetBytes($Secret))
    try {
        $signature = ConvertTo-Base64Url ($hmac.ComputeHash([Text.Encoding]::UTF8.GetBytes($signingInput)))
    }
    finally {
        $hmac.Dispose()
    }
    return "$signingInput.$signature"
}

function Register-TestSession {
    param([string]$UserID, [string]$Email, [string]$SessionID)
    $expiresAt = [DateTimeOffset]::UtcNow.AddHours(1).UtcDateTime.ToString("o")
    $payload = @{
        id         = $SessionID
        user_id    = $UserID
        email      = $Email
        expires_at = $expiresAt
    } | ConvertTo-Json -Compress

    $redisKey = "auth:sessions:$SessionID"
    $redisResult = $payload | docker exec -i snp-redis redis-cli -x SET $redisKey
    Assert-True ($LASTEXITCODE -eq 0) "Failed to seed Redis session for $UserID"
    Assert-True (($redisResult | Select-Object -Last 1) -eq "OK") "Unexpected Redis response while seeding session: $redisResult"
    docker exec snp-redis redis-cli EXPIRE $redisKey 3600 | Out-Null
    Assert-True ($LASTEXITCODE -eq 0) "Failed to set Redis session TTL for $UserID"
}

function Invoke-JsonRequest {
    param(
        [string]$Method,
        [string]$Path,
        [string]$Token,
        [object]$Body,
        [int[]]$ExpectedStatus = @(200)
    )

    $headers = @{
        "X-Request-ID"     = "load-setup-$([Guid]::NewGuid().ToString("N"))"
        "X-Correlation-ID" = "load-setup"
    }
    if ($Token) {
        $headers.Authorization = "Bearer $Token"
    }
    $bodyJson = $null
    if ($null -ne $Body) {
        $bodyJson = $Body | ConvertTo-Json -Compress
    }

    try {
        $response = Invoke-WebRequest -Uri "$GatewayBaseUrl$Path" -Method $Method -Headers $headers -Body $bodyJson -ContentType "application/json" -UseBasicParsing
        $status = [int]$response.StatusCode
        $content = $response.Content
    }
    catch {
        if ($_.Exception.Response -eq $null) {
            throw
        }
        $status = [int]$_.Exception.Response.StatusCode
        $stream = $_.Exception.Response.GetResponseStream()
        $reader = [IO.StreamReader]::new($stream)
        try {
            $content = $reader.ReadToEnd()
        }
        finally {
            $reader.Dispose()
        }
    }

    Assert-True ($ExpectedStatus -contains $status) "Expected $Method $Path status $($ExpectedStatus -join ',') but got $status. Body: $content"
    if ([string]::IsNullOrWhiteSpace($content)) {
        return [pscustomobject]@{ status = $status; body = $null }
    }
    return [pscustomobject]@{ status = $status; body = ($content | ConvertFrom-Json) }
}

function Wait-Until {
    param([scriptblock]$Condition, [string]$Description, [int]$Seconds = 180)
    $deadline = [DateTimeOffset]::UtcNow.AddSeconds($Seconds)
    $lastError = $null
    do {
        try {
            if (& $Condition) {
                return
            }
        }
        catch {
            $lastError = $_.Exception.Message
        }
        Start-Sleep -Seconds 2
    } while ([DateTimeOffset]::UtcNow -lt $deadline)
    if ($lastError) {
        throw "Timed out waiting for $Description. Last error: $lastError"
    }
    throw "Timed out waiting for $Description."
}

function Wait-Health {
    param(
        [string]$Url = "$GatewayBaseUrl/health",
        [string]$Description = "gateway health"
    )

    Wait-Until -Description $Description -Condition {
        $response = Invoke-WebRequest -Uri $Url -UseBasicParsing
        return [int]$response.StatusCode -eq 200
    }
}

function Wait-StackHealth {
    Wait-Health -Url "$GatewayBaseUrl/health" -Description "api-gateway health"
    Wait-Health -Url "http://localhost:8081/health" -Description "auth-service health"
    Wait-Health -Url "http://localhost:8082/health" -Description "users-service health"
    Wait-Health -Url "http://localhost:8083/health" -Description "posts-service health"
    Wait-Health -Url "http://localhost:8084/health" -Description "feed-service health"
    Wait-Health -Url "http://localhost:8085/health" -Description "notification-service health"
}

function New-LoadUser {
    param([string]$Prefix, [int]$Index, [string]$RunID)
    $userID = "$Prefix-$RunID-$Index"
    $email = "$userID@example.com"
    $sessionID = "session-$userID"
    $token = New-TestJwt -UserID $userID -Email $email -SessionID $sessionID -Secret $JwtSecret
    Register-TestSession -UserID $userID -Email $email -SessionID $sessionID
    Wait-Until -Description "gateway users-service readiness for $userID" -Seconds 60 -Condition {
        try {
            Invoke-JsonRequest -Method GET -Path "/api/v1/users/me" -Token $token -ExpectedStatus @(200) | Out-Null
            return $true
        }
        catch {
            Write-Host "Waiting for users-service setup path for ${userID}: $($_.Exception.Message)"
            return $false
        }
    }
    return [pscustomobject]@{
        UserID = $userID
        Email = $email
        SessionID = $sessionID
        Token = $token
    }
}

function Run-K6 {
    param(
        [string]$Script,
        [string]$SummaryPath,
        [string[]]$Arguments
    )

    $k6 = Get-Command k6 -ErrorAction SilentlyContinue
    if ($K6Runner -eq "local" -and -not $k6) {
        throw "k6 is not available in PATH. Install k6 or use -K6Runner docker."
    }

    if ($k6 -and $K6Runner -ne "docker") {
        & k6 run --summary-export $SummaryPath @Arguments $Script
        Assert-True ($LASTEXITCODE -eq 0) "k6 failed for $Script"
        return
    }

    $docker = Get-Command docker -ErrorAction SilentlyContinue
    if (-not $docker) {
        throw "k6 is not available in PATH and Docker is not available for the fallback runner."
    }

    $containerScript = Convert-ToContainerPath -Path $Script
    $containerSummary = Convert-ToContainerPath -Path $SummaryPath
    $dockerArgs = Convert-K6ArgsForDocker -Arguments $Arguments

    Write-Host "k6 CLI not found; running $K6DockerImage with Docker."
    $mount = "${RepoRoot}:/workspace"
    & docker run --rm `
        -v $mount `
        -w /workspace `
        $K6DockerImage `
        run --summary-export $containerSummary @dockerArgs $containerScript
    Assert-True ($LASTEXITCODE -eq 0) "k6 failed for $Script"
}

function Convert-ToContainerPath {
    param([string]$Path)
    $rootPath = (Resolve-Path $RepoRoot).Path.TrimEnd("\")
    if (Test-Path $Path) {
        $fullPath = (Resolve-Path $Path).Path
    }
    else {
        $parent = Split-Path -Parent $Path
        $leaf = Split-Path -Leaf $Path
        $parentPath = (Resolve-Path $parent).Path
        $fullPath = Join-Path $parentPath $leaf
    }
    Assert-True ($fullPath.StartsWith($rootPath)) "Path is outside repository root: $fullPath"
    $relative = $fullPath.Substring($rootPath.Length).TrimStart("\")
    return "/workspace/" + ($relative -replace "\\", "/")
}

function Convert-BaseUrlForDocker {
    param([string]$Value)
    return $Value -replace "^http://localhost", "http://host.docker.internal" -replace "^http://127\.0\.0\.1", "http://host.docker.internal"
}

function Convert-K6ArgsForDocker {
    param([string[]]$Arguments)
    $converted = @()
    for ($i = 0; $i -lt $Arguments.Count; $i++) {
        $arg = $Arguments[$i]
        if ($arg -eq "--env" -and ($i + 1) -lt $Arguments.Count) {
            $converted += $arg
            $next = $Arguments[$i + 1]
            if ($next -like "BASE_URL=*") {
                $baseUrl = $next.Substring("BASE_URL=".Length)
                $converted += "BASE_URL=$(Convert-BaseUrlForDocker -Value $baseUrl)"
            }
            else {
                $converted += $next
            }
            $i++
            continue
        }
        $converted += $arg
    }
    return $converted
}

function New-MarkdownReport {
    param(
        [string]$ReportPath,
        [string]$ReadSummary,
        [string]$WriteSummary,
        [object[]]$ReadUsers,
        [object[]]$ActorUsers,
        [string]$AuthorID,
        [string]$PostID
    )

    $generatedAt = [DateTimeOffset]::UtcNow.ToString("o")
    $lines = @(
        "# Issue 64 Load Test Report",
        "",
        "Generated at: $generatedAt",
        "",
        "| Scenario | Script | Summary JSON |",
        "| --- | --- | --- |"
    )
    if ($ReadSummary) {
        $lines += "| Gateway read load | tests/load/k6/gateway-read-load.js | $ReadSummary |"
    }
    if ($WriteSummary) {
        $lines += "| Social write stress | tests/load/k6/social-write-stress.js | $WriteSummary |"
    }
    $lines += @(
        "",
        "## Test Data",
        "",
        "- Read users seeded: $($ReadUsers.Count)",
        "- Actor users seeded: $($ActorUsers.Count)",
        "- Author user: $AuthorID",
        "- Seed post id: $PostID",
        "",
        "## Commands",
        "",
        "```powershell",
        "powershell -ExecutionPolicy Bypass -File tests\load\run-load-tests.ps1 -StartStack",
        "```",
        "",
        "## Notes",
        "",
        "- k6 threshold results are recorded in the JSON summary files listed above.",
        "- Use Grafana/Prometheus dashboards during the run to correlate service latency, error rate, Redis, Kafka, and database behavior."
    )
    $lines | Set-Content -LiteralPath $ReportPath -Encoding UTF8
}

if ($ValidateOnly) {
    Assert-True (Test-Path $ComposeFile) "Compose file not found: $ComposeFile"
    Assert-True (Test-Path $readScript) "Read load script not found: $readScript"
    Assert-True (Test-Path $writeScript) "Write stress script not found: $writeScript"
    Assert-True ($ReadVus -gt 0) "ReadVus must be greater than 0"
    Assert-True ($WriteVus -gt 0) "WriteVus must be greater than 0"
    Write-Host "load test script validation passed"
    exit 0
}

New-Item -ItemType Directory -Force -Path $reportsDir | Out-Null
Push-Location $RepoRoot
try {
    if ($StartStack) {
        & docker compose -f $ComposeFile up -d --build
        Assert-True ($LASTEXITCODE -eq 0) "docker compose up failed"
    }

    Wait-StackHealth

    $runID = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
    $readUsers = @()
    $actorUsers = @()
    $readUserCount = [Math]::Max($ReadVus * 3, 1)
    $actorUserCount = [Math]::Max($WriteVus, 1)

    for ($i = 1; $i -le $readUserCount; $i++) {
        $readUsers += New-LoadUser -Prefix "load-read" -Index $i -RunID $runID
    }
    $author = New-LoadUser -Prefix "load-author" -Index 1 -RunID $runID
    for ($i = 1; $i -le $actorUserCount; $i++) {
        $actorUsers += New-LoadUser -Prefix "load-actor" -Index $i -RunID $runID
    }

    $seedPost = Invoke-JsonRequest -Method POST -Path "/api/v1/posts" -Token $author.Token -Body @{
        content = "load test seed post $runID"
    } -ExpectedStatus @(201)
    $postID = $seedPost.body.data.id
    Assert-True (-not [string]::IsNullOrWhiteSpace($postID)) "Seed post creation did not return an id"

    $timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
    $readSummary = $null
    $writeSummary = $null

    if (-not $SkipRead) {
        $readSummary = Join-Path $reportsDir "gateway-read-load-$timestamp.json"
        $readTokens = ($readUsers | ForEach-Object { $_.Token }) -join ";"
        Run-K6 -Script $readScript -SummaryPath $readSummary -Arguments @(
            "--env", "BASE_URL=$GatewayBaseUrl",
            "--env", "TOKENS=$readTokens",
            "--env", "VUS=$ReadVus",
            "--env", "DURATION=$ReadDuration"
        )
    }

    if (-not $SkipWrite) {
        $writeSummary = Join-Path $reportsDir "social-write-stress-$timestamp.json"
        $actorTokens = ($actorUsers | ForEach-Object { $_.Token }) -join ";"
        Run-K6 -Script $writeScript -SummaryPath $writeSummary -Arguments @(
            "--env", "BASE_URL=$GatewayBaseUrl",
            "--env", "AUTHOR_TOKEN=$($author.Token)",
            "--env", "ACTOR_TOKENS=$actorTokens",
            "--env", "POST_ID=$postID",
            "--env", "VUS=$WriteVus",
            "--env", "DURATION=$WriteDuration"
        )
    }

    $reportPath = Join-Path $reportsDir "issue-64-load-report-$timestamp.md"
    New-MarkdownReport -ReportPath $reportPath -ReadSummary $readSummary -WriteSummary $writeSummary -ReadUsers $readUsers -ActorUsers $actorUsers -AuthorID $author.UserID -PostID $postID
    Write-Host "load tests passed"
    Write-Host "Wrote report: $reportPath"
}
finally {
    Pop-Location
}
