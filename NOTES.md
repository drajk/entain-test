# Notes for Reviewer

Hey, quick summary of what I did and where I would take this next.

## What Changed

All 5 tasks done. Kept changes targeted and minimal, did not want to over engineer.
Added a `Makefile` and VS Code launch configs for local dev workflow, usage steps in README.

## Design Decisions

- **REST routes**: racing kept the original `POST /v1/list-races` to not break the existing contract. Sports uses `POST /v1/sports/events` which is more RESTful, showed both styles intentionally.
- **Sorting on sports**: hardcoded to `advertised_start_time ASC`. Didn't add sort_by/sort_direction like racing since the task asked for minimal. Easy to port over if needed.
- **Status derivation**: computed on read rather than stored. No stale data, no background jobs. Trades off a tiny bit of compute per query but for this scale its totally fine.
- **SQL injection**: sort field is validated against an allowlist before being interpolated. Small thing but important.
- **AI usage**: used AI as a pairing tool, not a code generator. All code is reviewed and trimmed by me. AI tends to add more than whats needed, kept an eye on scope throughout.

## What I'd Do Next (Production Readiness)

This is a code test so I kept things scoped, but heres what I would want before shipping for real:

**CI/CD**

- Pipeline: lint > test > build > deploy
- Honestly CICD is the first thing I set up in any new codebase. Deploy a minimal slice to dev early so the team can iterate fast.
- Proto breaking change detection with `buf`

**Testing**

- Integration tests against real SQLite (or Postgres), treat the APIs and gateways as a blackbox, fast feedback loop
- Smoke tests for deploy verification, nothing fancy
- Would keep it Go native, no extra test framework overhead

**Observability**

- Structured logging (`zap` or `slog`)
- Metrics via Prometheus, request latency, error rates, DB query times
- Distributed tracing with OpenTelemetry

**Infrastructure**

- Dockerfiles for local dev and deployment
- Terraform or similar for infra provisioning
- Monorepo deploy strategy, detect changed services, deploy only whats needed

**Other bits**

- Go version bump (currently 1.16, would move to latest)
- OpenAPI spec generation from protos, `protoc-gen-openapiv2` is already a dep, just not wired up
- `context` propagation, pass request ctx through to DB calls (`QueryContext` instead of `Query`)
- Make racing routes more RESTful (`/v1/races` instead of `/v1/list-races`) if we are ok breaking the API
- Pagination on list endpoints, not needed now but will be eventually
- Graceful shutdown on the gRPC servers
