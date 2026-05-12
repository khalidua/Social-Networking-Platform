# GitHub Copilot Repository Instructions

This is a Go microservices Social Networking Platform. Follow the repository workflow in `docs/AGENT_RULES.md`.

Before suggesting implementation changes:

- Prefer existing patterns in the affected service.
- Preserve service boundaries and clean layering.
- Use standardized JSON response helpers and app error codes where present.
- Preserve request/correlation ID propagation, structured logging, and gateway auth/session/rate-limit behavior.
- Add or update tests for changed behavior.
- Update `docs/CHANGELOG_AI.md`, relevant `docs/context/*.md`, and implementation notes after completed work.

Do not modify `docs/AGENT_RULES.md` or other governance files unless explicitly requested.
