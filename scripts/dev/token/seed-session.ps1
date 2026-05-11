param(
    [string]$SessionId = "session-1",
    [string]$UserId = "google:123",
    [string]$Email = "user@example.com",
    [int]$TtlSeconds = 3600,
    [string]$RedisContainer = "snp-redis"
)

$ErrorActionPreference = "Stop"

$expiresAt = (Get-Date).ToUniversalTime().AddSeconds($TtlSeconds).ToString("yyyy-MM-ddTHH:mm:ssZ")
$payloadObj = @{
    id         = $SessionId
    user_id    = $UserId
    email      = $Email
    expires_at = $expiresAt
}
$payload = $payloadObj | ConvertTo-Json -Compress
$redisKey = "auth:sessions:$SessionId"

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    throw "docker is not available in PATH."
}

$payload | docker exec -i $RedisContainer redis-cli -x SET $redisKey | Out-Null
docker exec $RedisContainer redis-cli EXPIRE $redisKey $TtlSeconds | Out-Null

Write-Host "Seeded session '$SessionId' for user '$UserId' (TTL: $TtlSeconds s, expires_at: $expiresAt)."
