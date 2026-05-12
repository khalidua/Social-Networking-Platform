# Contract Tests

This folder contains durable contract tests for issue 10.4.

## What Is Tested

- API Gateway OpenAPI contract exists and includes required paths, methods, and schemas.
- Kafka event schema JSON files are valid JSON.
- Event schemas keep required fields for:
  - `user.followed`
  - `post.created`
  - `post.interacted`
- Event contract docs reference the active topics and schema files.

## Why It Matters

These tests catch incompatible API or event contract changes before they break gateway/service or producer/consumer compatibility.

## Run

From the repository root:

```powershell
powershell -ExecutionPolicy Bypass -File tests\contract\contract-validation.ps1
```

## Expected Result

```text
contract validation passed
```

## Edge Cases Covered

- Missing OpenAPI path or HTTP method.
- Missing required component schema.
- Invalid JSON schema files.
- Missing required event fields.
- Event schemas allowing undeclared additional properties.
- Event docs no longer referencing active topics/schema files.

## Failure Scenarios Covered

- Removing a gateway route from OpenAPI.
- Renaming an event payload field without updating the contract.
- Accidentally allowing extra event fields in strict event contracts.
- Deleting an event schema doc reference.
