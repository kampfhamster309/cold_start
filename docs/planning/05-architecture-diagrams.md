# cold_start — Architecture Diagrams

> Companion to [03-architecture.md](03-architecture.md) — every diagram here is a direct visualization of a section there; component names, module names, and flow steps match exactly so the two stay easy to cross-check. Where a diagram and the prose ever disagree, the prose in 03-architecture.md is authoritative and this file has drifted and needs fixing.

## 1. System Context

Matches architecture §1.

```mermaid
flowchart TB
    Browser["Browser<br/>(new hire, manager, HR staff)"]
    IdP["Identity Provider<br/>(OIDC/SSO, company IdP)"]
    Caddy["caddy<br/>(TLS termination, reverse proxy)"]
    Frontend["app-frontend<br/>(Next.js)"]
    Backend["app-backend (Go)"]
    Postgres[("postgres<br/>(data + MVP search)")]
    DocStore[("go-git bare repo(s)<br/>docstore-data volume")]
    Scratch[("scratch clones<br/>ingestion-scratch volume")]
    CodeHost["Code hosts<br/>(GitLab self-managed first,<br/>GitHub.com / self-hosted git later)"]
    SMTP["Operator-supplied SMTP relay"]

    Browser -- HTTPS --> Caddy
    Browser -- login redirect --> IdP
    Caddy --> Frontend
    Frontend -- REST/JSON --> Backend
    Backend -- reads/writes --> Postgres
    Backend -- git blob/tree/commit ops --> DocStore
    Backend -- clone + blame/churn --> Scratch
    Backend -- REST/GraphQL + webhooks --> CodeHost
    Backend -- magic-link delivery --> SMTP
```

## 2. Backend Component Breakdown

Matches architecture §2.2 — `app-backend` is a single binary/service at MVP, internally modularized. Solid arrows are data/control flow; dashed arrows are permission checks (every write and read path routes through Auth/RBAC before touching its target).

```mermaid
flowchart LR
    subgraph Backend["app-backend (Go)"]
        API["API layer"]
        Auth["Auth/RBAC module"]
        DocEngine["Doc-store engine"]
        Ingestion["Ingestion workers"]
        Planner["Onboarding planner module"]
        Search["Search module"]
    end

    API --> Auth
    API --> DocEngine
    API --> Planner
    API --> Search

    Auth -.grant check.-> DocEngine
    Auth -.grant check.-> Search
    Auth -.grant check.-> Planner

    DocEngine --> DocStoreVol[("go-git bare repo(s)")]
    Ingestion --> ScratchVol[("scratch clones")]
    Ingestion --> PG[("postgres")]
    DocEngine --> PG
    Planner --> PG
    Search --> PG
    Auth --> PG
```

## 3. Deployment Topology

Matches architecture §4.

```mermaid
flowchart TB
    subgraph Compose["docker-compose — one instance per company, self-hosted"]
        Caddy["caddy"]
        FE["app-frontend"]
        BE["app-backend"]
        PG["postgres"]
        Backup["backup (scheduled cron, not long-running)"]
    end

    PGData[("pg-data volume")]
    DocData[("docstore-data volume")]
    ScratchData[("ingestion-scratch volume")]
    OffHost[("operator off-host storage")]
    SMTPExt["Operator-supplied SMTP relay (external)"]

    Caddy --> FE
    Caddy --> BE
    BE --> DocData
    BE --> ScratchData
    PG --> PGData
    BE --> SMTPExt
    Backup -- coordinated snapshot --> PGData
    Backup -- coordinated snapshot --> DocData
    Backup --> OffHost
```

## 4. Sequence: Repository Ingestion → "Who to Ask" (Phase 0) / Staleness Signals & Repo-Doc Search (Phase 1)

Matches architecture §3.1. Phase boundary marked inline — the Phase 1 steps are additive, not a rework of the Phase 0 loop.

```mermaid
sequenceDiagram
    actor Admin
    participant FE as app-frontend
    participant BE as app-backend
    participant PG as postgres
    participant Host as Code host
    participant Clone as scratch volume

    Admin->>FE: Register repo connection
    FE->>BE: POST repo connection
    BE->>PG: Store connection (host type, token)
    BE->>BE: Schedule polling (default trigger)
    opt operator confirmed reachable ingress
        BE->>Host: Register webhook
    end
    loop poll tick (or webhook delivery, where enabled)
        BE->>Host: Fetch commits / PR / review metadata
        BE->>Clone: Refresh full (non-shallow) local clone
        BE->>Clone: git blame / git log --stat
        BE->>PG: Write per-path expertise scores [Phase 0]
        opt Phase 1
            BE->>Clone: Read configured doc paths (README, /docs)
            BE->>PG: Write repo-doc content to search corpus<br/>(resource_type=repo_connection)
            BE->>PG: Compare churn vs. doc freshness,<br/>write staleness flags
        end
    end
    actor User
    User->>FE: View repo dashboard
    FE->>BE: GET precomputed signals
    BE->>PG: Read signals
    PG-->>BE: signals
    BE-->>FE: signals (no live git/API call on this path)
```

## 5. Sequence: Non-Code Doc Edit

Matches architecture §3.2.

```mermaid
sequenceDiagram
    actor User
    participant FE as app-frontend
    participant BE as app-backend
    participant Git as go-git bare repo
    participant PG as postgres

    User->>FE: Open doc page
    FE->>BE: GET content + base commit hash
    BE->>Git: Read HEAD content
    Git-->>BE: content, hash
    BE-->>FE: content, base hash
    User->>FE: Edit and save
    FE->>BE: PUT new content + base hash
    BE->>Git: Compare base hash to current HEAD
    alt base hash stale (concurrent edit)
        BE-->>FE: 409 conflict — reload to see latest
    else base hash current
        BE->>Git: Write blob/tree/commit, update ref
        BE->>PG: Update metadata + search-index row (same request)
        BE-->>FE: 200 saved
    end
    User->>FE: View history / diff / rollback
    FE->>BE: GET history
    BE->>Git: git log / git diff (not reconstructed from Postgres)
    Git-->>BE: history/diff
    BE-->>FE: history/diff
```

## 6. Sequence: Onboarding Plan Assignment

Matches architecture §3.3.

```mermaid
sequenceDiagram
    actor Manager as Manager/buddy
    participant FE as app-frontend
    participant BE as app-backend
    participant PG as postgres
    actor Hire as New hire

    Manager->>FE: Create/select plan template
    FE->>BE: POST template (CRUD)
    BE->>PG: Store template
    Manager->>FE: Assign plan to new hire
    FE->>BE: POST assignment
    BE->>PG: Instantiate per-hire step records<br/>(doc-store ID, repo ID, task, contact, link)
    BE->>PG: Issue grant (manager, viewer, onboarding_hire, hire_id)
    BE-->>FE: assignment created
    Hire->>FE: View progress
    FE->>BE: GET progress
    BE->>PG: Read steps, filtered by caller's grants
    PG-->>BE: steps + completion state
    BE-->>FE: progress view
    Hire->>FE: Mark step complete
    FE->>BE: PATCH step
    alt prior step (N) incomplete
        BE-->>FE: 409 rejected — step N+1 locked
    else prior step complete
        BE->>PG: Write completed_at (+ started_at if unset)
        BE-->>FE: 200 updated
    end
```

## 7. Sequence: Auth

Matches architecture §3.4.

```mermaid
sequenceDiagram
    actor User
    participant FE as app-frontend
    participant BE as app-backend
    participant IdP as Company OIDC IdP
    participant SMTP as SMTP relay
    participant PG as postgres

    alt OIDC login
        User->>FE: Sign in
        FE->>IdP: Redirect
        IdP-->>BE: OIDC callback (asserted email)
    else Magic link (account outside IdP tenant)
        User->>FE: Request magic link
        BE->>SMTP: Send magic-link email
        User->>BE: Follow link
    end
    BE->>PG: Look up existing account by email
    alt email matches an existing magic-link account
        BE-->>User: Require explicit confirmation before linking<br/>(no silent auto-merge)
    else no conflict
        BE->>PG: Create/update user row
        BE-->>FE: Issue session cookie
    end
    User->>FE: Any subsequent request
    FE->>BE: Request + session cookie
    BE->>PG: Check (user, role, resource_type, resource_id) grants
    PG-->>BE: grants
    BE-->>FE: response, filtered by grants
```
