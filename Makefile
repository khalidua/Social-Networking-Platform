SERVICES := api-gateway auth-service users-service posts-service feed-service notification-service
GOCACHE  ?= $(CURDIR)/.gocache
export GOCACHE
export GOTELEMETRY=off

.PHONY: help test test-unit test-integration test-contract test-load-validate \
        coverage build vet clean ci

# ── Help ──────────────────────────────────────────────────────────────────────
help:
	@echo ""
	@echo "Usage: make <target>"
	@echo ""
	@echo "  test                  Run all unit + in-memory integration tests"
	@echo "  test-unit             Run unit tests for all services"
	@echo "  test-integration      Run in-memory integration tests (posts + users)"
	@echo "  test-contract         Validate OpenAPI spec and event schemas"
	@echo "  test-load-validate    Validate k6 load/stress scripts (syntax only)"
	@echo "  coverage              Print coverage summary for all service layers"
	@echo "  build                 Build binaries for all services"
	@echo "  vet                   Run go vet on all services"
	@echo "  clean                 Remove build artifacts and coverage reports"
	@echo "  ci                    Run the full local CI gate (unit + integration + contract)"
	@echo ""

# ── Unit tests ────────────────────────────────────────────────────────────────
test-unit:
	@for svc in $(SERVICES); do \
	  echo "=== Unit tests: $$svc ==="; \
	  (cd $$svc && go test -race -covermode=atomic -coverprofile=coverage.out \
	    $$(go list ./... | grep -v '/integration$$') ./...) || exit 1; \
	done

# ── In-memory integration tests ───────────────────────────────────────────────
test-integration:
	@echo "=== Integration tests: posts-service (in-memory) ==="
	(cd posts-service && go test -v -race ./internal/integration/...)
	@echo ""
	@echo "=== Integration tests: users-service (requires INTEGRATION_PG_DSN) ==="
	@if [ -z "$$INTEGRATION_PG_DSN" ]; then \
	  echo "  SKIPPED: set INTEGRATION_PG_DSN to run Postgres integration tests"; \
	else \
	  (cd users-service && go test -v -race -tags integration ./internal/integration/...); \
	fi

# ── Contract tests ────────────────────────────────────────────────────────────
test-contract:
	@echo "=== Contract validation ==="
	@python3 - <<'EOF'
	import json, sys, pathlib, re

	errors = []

	# --- Event schemas ---
	schemas_dir = pathlib.Path("docs/schemas")
	required_fields = {
	    "user-followed-v1.json":   ["follower_id", "followee_id"],
	    "post-created-v1.json":    ["post_id", "author_id", "content", "created_at"],
	    "post-interacted-v1.json": ["post_id", "post_author_id", "actor_id",
	                                "interaction_type", "created_at"],
	}
	for fname, fields in required_fields.items():
	    path = schemas_dir / fname
	    if not path.exists():
	        errors.append(f"Missing: {path}"); continue
	    schema = json.loads(path.read_text())
	    for f in fields:
	        if f not in schema.get("required", []):
	            errors.append(f"{fname}: missing required field '{f}'")
	        if f not in schema.get("properties", {}):
	            errors.append(f"{fname}: missing property '{f}'")
	    if schema.get("additionalProperties") is not False:
	        errors.append(f"{fname}: must set additionalProperties=false")

	# --- Event contracts doc ---
	doc = pathlib.Path("docs/messaging/event-contracts.md").read_text()
	for item in ["user.followed","post.created","post.interacted",
	             "user-followed-v1.json","post-created-v1.json","post-interacted-v1.json"]:
	    if item not in doc:
	        errors.append(f"event-contracts.md missing '{item}'")

	# --- OpenAPI ---
	oas = pathlib.Path("docs/openapi/swagger.yaml").read_text()
	for name in ["SuccessEnvelope","ErrorEnvelope","AuthUser","SessionValidationResponse",
	             "User","Post","PostInteraction","Notification"]:
	    if f"    {name}:" not in oas:
	        errors.append(f"OpenAPI missing schema '{name}'")

	if errors:
	    for e in errors: print("FAIL:", e)
	    sys.exit(1)
	print("PASS: all contract checks passed")
	EOF

# ── Load script validation ─────────────────────────────────────────────────────
test-load-validate:
	@echo "=== k6 script validation ==="
	@which k6 > /dev/null 2>&1 || (echo "  k6 not found; skipping syntax check" && exit 0)
	k6 inspect tests/load/k6/gateway-read-load.js
	k6 inspect tests/load/k6/social-write-stress.js
	@echo "PASS: k6 scripts are valid"

# ── Coverage summary ──────────────────────────────────────────────────────────
coverage: test-unit
	@echo ""
	@echo "=== Coverage summary (service layer) ==="
	@for svc in $(SERVICES); do \
	  if [ -f $$svc/coverage.out ]; then \
	    echo "--- $$svc ---"; \
	    (cd $$svc && go tool cover -func=coverage.out | grep -E "internal/service|^total"); \
	  fi; \
	done

# ── Build ─────────────────────────────────────────────────────────────────────
build:
	@for svc in $(SERVICES); do \
	  echo "Building $$svc..."; \
	  (cd $$svc && go build ./cmd/server/...) || exit 1; \
	done
	@echo "PASS: all services compiled"

# ── Vet ───────────────────────────────────────────────────────────────────────
vet:
	@for svc in $(SERVICES); do \
	  echo "Vetting $$svc..."; \
	  (cd $$svc && go vet ./...) || exit 1; \
	done
	@echo "PASS: go vet clean"

# ── Full local CI gate ─────────────────────────────────────────────────────────
ci: vet test-unit test-integration test-contract test-load-validate build
	@echo ""
	@echo "✓ Local CI gate passed"

# ── Clean ─────────────────────────────────────────────────────────────────────
clean:
	@for svc in $(SERVICES); do \
	  rm -f $$svc/coverage.out $$svc/coverage-summary.txt $$svc/test-output.txt; \
	done
	@echo "Cleaned coverage and test artifacts"