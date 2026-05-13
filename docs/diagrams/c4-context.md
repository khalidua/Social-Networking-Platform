# C4 Context View

```mermaid
flowchart LR
    user["User / Evaluator"]
    google["Google OAuth2"]
    platform["Social Networking Platform"]

    user -->|"Browser / API client"| platform
    platform -->|"OAuth authorization code flow"| google
```

## Notes

- The API Gateway is the public entry point for the platform.
- Google OAuth2 is the only external identity provider.
- Local automated tests can use seeded JWT and Redis sessions to avoid external Google dependency.
