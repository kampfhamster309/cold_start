# Onboarding Hub

A self-hosted, web-based onboarding platform for software engineering teams. It connects to code repositories to surface "who to ask" expertise and documentation-staleness signals, provides a git-backed store for non-code documentation editable from a web UI, and includes an onboarding-plan/checklist feature for new hires.

**Status: early scaffolding.** Only the base backend/frontend services and their health-check round trip exist so far (Phase 0, tickets `INFRA-1`/`INFRA-2`). See [docs/planning/06-development-plan.md](docs/planning/06-development-plan.md) for what's next.

## Documentation

The full product vision, technology decisions, and system architecture live in [`docs/planning/`](docs/planning) — start with [`01-overview.md`](docs/planning/01-overview.md). That directory is the source of truth for *why* things are built the way they are; this README is just a front door.

| Doc | Contents |
|---|---|
| [01-overview.md](docs/planning/01-overview.md) | Vision, goals, personas, feature scope, phasing |
| [02-tech-stack.md](docs/planning/02-tech-stack.md) | Technology stack and the reasoning behind each pick |
| [03-architecture.md](docs/planning/03-architecture.md) | Components, data flows, deployment topology, security model |
| [04-open-items.md](docs/planning/04-open-items.md) | Decisions still needed vs. deliberately deferred |
| [05-architecture-diagrams.md](docs/planning/05-architecture-diagrams.md) | Diagrams matching the architecture doc |
| [06-08] development plans | Phase 0/1/2 broken into epics and implementation tickets |

## Project layout

```
backend/    Go app-backend (API, auth, doc-store, ingestion, planner, search — architecture §2.2)
frontend/   Next.js app-frontend (App Router, TypeScript)
docs/       Planning documentation (see above)
```

## Getting started

Requires Docker.

```bash
docker compose up -d --build
curl http://localhost:8080/healthz   # backend health check
open http://localhost:3000           # frontend, round-trips a call to the backend
docker compose down
```

For backend-only local development (no Docker): see the commands in `CLAUDE.md`.

## Tech stack

Go backend (single static binary), Next.js frontend, PostgreSQL, `go-git`-managed doc storage, GitLab self-managed as the first code-host integration, docker-compose for self-hosted deployment. Full reasoning in [02-tech-stack.md](docs/planning/02-tech-stack.md).
