# Issue 64 Baseline Load Test Report

## Scope

Issue #64 requires runnable load/stress scripts and recorded results for concurrency/scalability evaluation.

## Scripts Added

| Scenario | Script | Status |
| --- | --- | --- |
| Gateway read load | `tests/load/k6/gateway-read-load.js` | Ready |
| Social write stress | `tests/load/k6/social-write-stress.js` | Ready |
| PowerShell runner | `tests/load/run-load-tests.ps1` | Ready |

## Validation Results

Run from repository root:

```powershell
powershell -ExecutionPolicy Bypass -File tests\load\run-load-tests.ps1 -ValidateOnly
```

Expected result:

```text
load test script validation passed
```

## Full Run Command

```powershell
powershell -ExecutionPolicy Bypass -File tests\load\run-load-tests.ps1 -StartStack
```

The full run writes timestamped k6 JSON summaries and a Markdown report under `tests/load/reports/`.

## Result Recording Notes

- Record the generated Markdown report path in the final project report.
- Attach the generated k6 JSON summaries if detailed metrics are needed.
- Correlate results with Prometheus/Grafana dashboards from issue 9.

## Current Local Execution Status

The scripts are repository-ready. Full performance numbers must be collected on a machine with Docker Desktop and k6 available, with no stale `snp-*` container-name conflicts.
