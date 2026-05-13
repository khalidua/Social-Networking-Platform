param(
    [string]$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..\..\..")).Path,
    [string]$ComposeFile = (Join-Path $RepoRoot "deploy\compose\compose.yml"),
    [string]$GatewayBaseUrl = "http://localhost:8080",
    [string]$JwtSecret = "change-me",
    [int]$TimeoutSeconds = 180,
    [switch]$StartStack,
    [switch]$ValidateOnly
)

$ErrorActionPreference = "Stop"

function Assert-True {
    param(
        [bool]$Condition,
        [string]$Message
    )
    if (-not $Condition) {
        throw $Message
    }
}

function ConvertTo-Base64Url {
    param([byte[]]$Bytes)
    return [Convert]::ToBase64String($Bytes).TrimEnd("=").Replace("+", "-").Replace("/", "_")
}

function New-TestJwt {
    param(
        [string]$UserID,
        [string]$Email,
        [string]$SessionID,
        [string]$Secret
    )

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

function Invoke-JsonRequest {
    param(
        [string]$Method,
        [string]$Path,
        [string]$Token,
        [object]$Body,
        [int[]]$ExpectedStatus = @(200)
    )

    $headers = @{
        "X-Request-ID"     = "e2e-$([Guid]::NewGuid().ToString("N"))"
        "X-Correlation-ID" = "e2e-user-flow"
    }
    if ($Token) {
        $headers.Authorization = "Bearer $Token"
    }

    $uri = "$GatewayBaseUrl$Path"
    $bodyJson = $null
    if ($null -ne $Body) {
        $bodyJson = $Body | ConvertTo-Json -Compress
    }

    try {
        $response = Invoke-WebRequest -Uri $uri -Method $Method -Headers $headers -Body $bodyJson -ContentType "application/json" -UseBasicParsing
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
    param(
        [scriptblock]$Condition,
        [string]$Description,
        [int]$Seconds = $TimeoutSeconds
    )

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
    param([string]$Path)

    Wait-Until -Description "healthy $Path" -Condition {
        $result = Invoke-JsonRequest -Method GET -Path $Path -ExpectedStatus @(200)
        return $result.body.status -eq "ok"
    }
}

function Wait-DirectHealth {
    param([string]$Url)

    Wait-Until -Description "healthy $Url" -Condition {
        $response = Invoke-WebRequest -Uri $Url -Method GET -UseBasicParsing
        if ([int]$response.StatusCode -ne 200) {
            return $false
        }
        $body = $response.Content | ConvertFrom-Json
        return $body.status -eq "ok"
    }
}

function Register-TestSession {
    param(
        [string]$UserID,
        [string]$Email,
        [string]$SessionID
    )

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

function Get-Items {
    param([object]$Value)
    if ($null -eq $Value) {
        return @()
    }
    if ($Value.PSObject.Properties.Name -contains "items") {
        return @($Value.items)
    }
    if ($Value.PSObject.Properties.Name -contains "Items") {
        return @($Value.Items)
    }
    return @()
}

if ($ValidateOnly) {
    Assert-True (Test-Path $ComposeFile) "Compose file not found: $ComposeFile"
    Assert-True ($GatewayBaseUrl.StartsWith("http")) "GatewayBaseUrl must be an HTTP URL"
    Write-Host "e2e user flow script validation passed"
    exit 0
}

Push-Location $RepoRoot
try {
    if ($StartStack) {
        & docker compose -f $ComposeFile up -d --build
        Assert-True ($LASTEXITCODE -eq 0) "docker compose up failed"
    }

    Wait-Health -Path "/health"
    Wait-DirectHealth -Url "http://localhost:8081/health"
    Wait-DirectHealth -Url "http://localhost:8082/health"
    Wait-DirectHealth -Url "http://localhost:8083/health"
    Wait-DirectHealth -Url "http://localhost:8084/health"
    Wait-DirectHealth -Url "http://localhost:8085/health"

    $suffix = [DateTimeOffset]::UtcNow.ToUnixTimeSeconds()
    $aliceID = "e2e-alice-$suffix"
    $bobID = "e2e-bob-$suffix"
    $aliceEmail = "$aliceID@example.com"
    $bobEmail = "$bobID@example.com"
    $aliceSession = "session-$aliceID"
    $bobSession = "session-$bobID"
    $aliceToken = New-TestJwt -UserID $aliceID -Email $aliceEmail -SessionID $aliceSession -Secret $JwtSecret
    $bobToken = New-TestJwt -UserID $bobID -Email $bobEmail -SessionID $bobSession -Secret $JwtSecret

    Register-TestSession -UserID $aliceID -Email $aliceEmail -SessionID $aliceSession
    Register-TestSession -UserID $bobID -Email $bobEmail -SessionID $bobSession

    Invoke-JsonRequest -Method GET -Path "/api/v1/users/me" -ExpectedStatus @(401) | Out-Null

    $sessionResult = Invoke-JsonRequest -Method GET -Path "/api/v1/auth/session" -Token $aliceToken -ExpectedStatus @(200)
    Assert-True ($sessionResult.body.success -eq $true) "Auth session validation did not return success"
    Assert-True ($sessionResult.body.data.user.id -eq $aliceID) "Auth session did not return the expected user id"

    $profileResult = Invoke-JsonRequest -Method PATCH -Path "/api/v1/users/me" -Token $aliceToken -Body @{
        name = "Alice Integration"
        bio  = "E2E test user"
    } -ExpectedStatus @(200)
    Assert-True ($profileResult.body.data.id -eq $aliceID) "Profile update did not return Alice"
    Assert-True ($profileResult.body.data.name -eq "Alice Integration") "Profile update did not persist name"

    Invoke-JsonRequest -Method GET -Path "/api/v1/users/me" -Token $bobToken -ExpectedStatus @(200) | Out-Null
    Invoke-JsonRequest -Method POST -Path "/api/v1/users/$bobID/follow" -Token $aliceToken -ExpectedStatus @(204) | Out-Null

    Wait-Until -Description "follow notification for Bob" -Seconds 60 -Condition {
        $notifications = Invoke-JsonRequest -Method GET -Path "/api/v1/notifications" -Token $bobToken -ExpectedStatus @(200)
        return @($notifications.body.data | Where-Object { $_.type -eq "follow" -and $_.message -like "*$aliceID*" }).Count -gt 0
    }

    $postResult = Invoke-JsonRequest -Method POST -Path "/api/v1/posts" -Token $bobToken -Body @{
        content = "hello from issue 62 e2e $suffix"
    } -ExpectedStatus @(201)
    $postID = $postResult.body.data.id
    Assert-True (-not [string]::IsNullOrWhiteSpace($postID)) "Post creation did not return an id"
    Assert-True ($postResult.body.data.authorId -eq $bobID) "Post author id mismatch"

    $getPostResult = Invoke-JsonRequest -Method GET -Path "/api/v1/posts/$postID" -Token $aliceToken -ExpectedStatus @(200)
    Assert-True ($getPostResult.body.data.id -eq $postID) "Get post did not return created post"

    $authorPosts = Invoke-JsonRequest -Method GET -Path "/api/v1/posts?authorId=$bobID" -Token $aliceToken -ExpectedStatus @(200)
    Assert-True (@($authorPosts.body.data | Where-Object { $_.id -eq $postID }).Count -gt 0) "List posts did not include created post"

    Wait-Until -Description "Alice feed contains Bob post" -Seconds 90 -Condition {
        $feed = Invoke-JsonRequest -Method GET -Path "/api/v1/feed" -Token $aliceToken -ExpectedStatus @(200)
        $items = Get-Items $feed.body.data
        return @($items | Where-Object { $_.PostID -eq $postID -or $_.post_id -eq $postID }).Count -gt 0
    }

    Invoke-JsonRequest -Method POST -Path "/api/v1/posts/$postID/interactions" -Token $aliceToken -Body @{
        interaction_type = "like"
    } -ExpectedStatus @(202) | Out-Null

    Wait-Until -Description "post like notification for Bob" -Seconds 60 -Condition {
        $notifications = Invoke-JsonRequest -Method GET -Path "/api/v1/notifications" -Token $bobToken -ExpectedStatus @(200)
        return @($notifications.body.data | Where-Object { $_.type -eq "post_like" -and $_.message -like "*$postID*" }).Count -gt 0
    }

    Write-Host "e2e user flow passed"
}
finally {
    Pop-Location
}
