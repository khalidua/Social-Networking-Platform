$ErrorActionPreference = "Stop"

$root = Resolve-Path (Join-Path $PSScriptRoot "..\..")

function Assert-FileExists {
    param([string] $Path)
    if (-not (Test-Path $Path)) {
        throw "Missing required contract file: $Path"
    }
}

function Assert-Contains {
    param(
        [string] $Content,
        [string] $Expected,
        [string] $Description
    )
    if (-not $Content.Contains($Expected)) {
        throw "Contract check failed: $Description. Missing '$Expected'."
    }
}

function Assert-JsonSchemaRequiredFields {
    param(
        [string] $Path,
        [string[]] $ExpectedRequired
    )

    Assert-FileExists $Path
    $schema = Get-Content -Raw -Path $Path | ConvertFrom-Json
    $actual = @($schema.required)
    foreach ($field in $ExpectedRequired) {
        if ($actual -notcontains $field) {
            throw "Contract check failed: '$Path' missing required field '$field'."
        }
    }
    foreach ($field in $ExpectedRequired) {
        if (-not $schema.properties.PSObject.Properties.Name.Contains($field)) {
            throw "Contract check failed: '$Path' missing property definition '$field'."
        }
    }
    if ($schema.additionalProperties -ne $false) {
        throw "Contract check failed: '$Path' must set additionalProperties to false."
    }
}

function Assert-OpenApiPathMethod {
    param(
        [string] $OpenApi,
        [string] $Path,
        [string] $Method
    )

    $pathPattern = "(?ms)^\s{2}$([regex]::Escape($Path)):\s*$"
    $match = [regex]::Match($OpenApi, $pathPattern)
    if (-not $match.Success) {
        throw "Contract check failed: OpenAPI missing path '$Path'."
    }

    $nextPath = [regex]::Match($OpenApi.Substring($match.Index + $match.Length), "(?m)^\s{2}/")
    $sectionLength = if ($nextPath.Success) { $nextPath.Index } else { $OpenApi.Length - ($match.Index + $match.Length) }
    $section = $OpenApi.Substring($match.Index + $match.Length, $sectionLength)
    if ($section -notmatch "(?m)^\s{4}$Method`:") {
        throw "Contract check failed: OpenAPI path '$Path' missing method '$Method'."
    }
}

$openApiPath = Join-Path $root "docs\openapi\swagger.yaml"
$eventContractsPath = Join-Path $root "docs\messaging\event-contracts.md"
$schemasDir = Join-Path $root "docs\schemas"

Assert-FileExists $openApiPath
Assert-FileExists $eventContractsPath

$openApi = Get-Content -Raw -Path $openApiPath
$eventContracts = Get-Content -Raw -Path $eventContractsPath

Assert-Contains $openApi "openapi: 3.0.0" "OpenAPI version"
foreach ($schemaName in @(
    "SuccessEnvelope",
    "ErrorEnvelope",
    "AuthUser",
    "SessionValidationResponse",
    "User",
    "Post",
    "PostInteraction",
    "Notification"
)) {
    Assert-Contains $openApi "    ${schemaName}:" "OpenAPI schema $schemaName"
}

$requiredPaths = @(
    @{ Path = "/health"; Method = "get" },
    @{ Path = "/api/v1/auth/login"; Method = "get" },
    @{ Path = "/api/v1/auth/callback"; Method = "get" },
    @{ Path = "/api/v1/auth/logout"; Method = "post" },
    @{ Path = "/api/v1/auth/session"; Method = "get" },
    @{ Path = "/api/v1/users/me"; Method = "get" },
    @{ Path = "/api/v1/users/me"; Method = "patch" },
    @{ Path = "/api/v1/users/{id}"; Method = "get" },
    @{ Path = "/api/v1/users/{id}/follow"; Method = "post" },
    @{ Path = "/api/v1/users/{id}/follow"; Method = "delete" },
    @{ Path = "/api/v1/posts"; Method = "post" },
    @{ Path = "/api/v1/posts/{id}"; Method = "get" },
    @{ Path = "/api/v1/posts/{id}"; Method = "patch" },
    @{ Path = "/api/v1/posts/{id}"; Method = "delete" },
    @{ Path = "/api/v1/posts/{id}/interactions"; Method = "post" },
    @{ Path = "/api/v1/feed"; Method = "get" },
    @{ Path = "/api/v1/notifications"; Method = "get" }
)

foreach ($entry in $requiredPaths) {
    Assert-OpenApiPathMethod -OpenApi $openApi -Path $entry.Path -Method $entry.Method
}

Assert-JsonSchemaRequiredFields `
    -Path (Join-Path $schemasDir "user-followed-v1.json") `
    -ExpectedRequired @("follower_id", "followee_id")

Assert-JsonSchemaRequiredFields `
    -Path (Join-Path $schemasDir "post-created-v1.json") `
    -ExpectedRequired @("post_id", "author_id", "content", "created_at")

Assert-JsonSchemaRequiredFields `
    -Path (Join-Path $schemasDir "post-interacted-v1.json") `
    -ExpectedRequired @("post_id", "post_author_id", "actor_id", "interaction_type", "created_at")

foreach ($topic in @("user.followed", "post.created", "post.interacted")) {
    Assert-Contains $eventContracts $topic "event topic $topic"
}
foreach ($schemaFile in @("user-followed-v1.json", "post-created-v1.json", "post-interacted-v1.json")) {
    Assert-Contains $eventContracts $schemaFile "event schema reference $schemaFile"
}

Write-Output "contract validation passed"
