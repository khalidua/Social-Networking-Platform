# Manual Test Plan

## Feature Being Tested

Issue 10.4 API and event contract tests.

## Preconditions

- PowerShell is available.
- Repository docs and schemas are present.

## Steps

1. From the repository root, run:

   ```powershell
   powershell -ExecutionPolicy Bypass -File tests\contract\contract-validation.ps1
   ```

2. Confirm the command exits successfully.

3. Optional negative check: temporarily remove a required field from one schema, run the command, confirm it fails, then restore the file.

## Expected Results

- The script prints `contract validation passed`.
- The script exits non-zero if an OpenAPI path/method, schema, event required field, or event doc reference is missing.

## Edge Cases

- Missing OpenAPI route.
- Missing OpenAPI schema.
- Invalid JSON event schema.
- Missing event required field.

## Failure Cases

- Breaking API compatibility by deleting a path or method.
- Breaking Kafka compatibility by renaming/removing event fields.
- Removing event contract documentation.

## Regression Checks

- Existing OpenAPI and Kafka event contracts remain compatible.
- Contract checks can be added to CI without requiring Docker.
