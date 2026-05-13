# Final Demo Flow

## 10-Minute Walkthrough

1. Start the stack:
   ```powershell
   powershell -ExecutionPolicy Bypass -File deploy\scripts\up.ps1 -Build
   ```
2. Verify health:
   ```powershell
   powershell -ExecutionPolicy Bypass -File deploy\scripts\health.ps1
   ```
3. Show auth options:
   - Real Google login through `http://localhost:8080/api/v1/auth/login` when credentials are configured.
   - Seeded JWT/Redis sessions through the E2E and load runners for deterministic evaluation.
4. Run the E2E user flow:
   ```powershell
   powershell -ExecutionPolicy Bypass -File tests\integration\e2e\e2e-user-flow.ps1
   ```
5. Show profile update, follow, post creation, feed retrieval, and notifications in the script output.
6. Open Grafana at `http://localhost:3000` and show:
   - request rate
   - p95/p99 latency
   - 5xx error rate
   - business operation rate
   - DB query rate and DB errors
7. Run load:
   ```powershell
   powershell -ExecutionPolicy Bypass -File tests\load\run-load-tests.ps1 -K6Runner docker -ReadVus 20 -ReadDuration "2m" -WriteVus 10 -WriteDuration "1m"
   ```
8. Enable latency simulation for the feed path:
   ```powershell
   $env:DEMO_SIMULATION_ENABLED="true"
   $env:DEMO_SIMULATION_PATH="/api/v1/feed"
   $env:DEMO_LATENCY="2s"
   $env:DEMO_FAILURE_RATE="0"
   docker compose -f deploy\compose\compose.yml up -d --build api-gateway
   ```
9. Enable failure simulation:
   ```powershell
   $env:DEMO_FAILURE_RATE="0.3"
   docker compose -f deploy\compose\compose.yml up -d --build api-gateway
   ```
10. Reset demo flags and restart gateway for normal behavior.

## Fallbacks

- If Google OAuth is unavailable, use the seeded-session E2E/load scripts.
- If local k6 is unavailable, use `-K6Runner docker`.
- If a container name conflict appears, run `deploy\scripts\down.ps1` and start again.
