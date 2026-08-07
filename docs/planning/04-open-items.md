# Onboarding Hub — Open Items

> Single tracker for everything still undecided or deliberately deferred across [01-overview.md](01-overview.md), [02-tech-stack.md](02-tech-stack.md), and [03-architecture.md](03-architecture.md). Those docs previously carried their own scattered "open questions" lists, which drifted out of sync with each other more than once during review — this file is now the only place that tracks status, so a decision only needs to be resolved in one place. The three docs still explain *why* each item exists; this file tracks *whether it's resolved yet*.

## 1. Decisions still needed

These need actual input (engineering judgment, legal/compliance review, or a product call) before the area they touch is either built or considered done. None of them block starting Phase 0 work overall — each note says specifically what it does block.

| Item | Status | Owner | Blocks |
|---|---|---|---|
| Data retention/anonymization policy for ingested commit-author data and per-hire onboarding-completion timing | Open — needs legal/compliance input, not resolvable by engineering judgment alone | Legal/compliance | Broad company rollout; not Phase 0 development itself, since the data is already scoped to be held "only as long as needed to compute aggregate signals" (architecture §5) as a placeholder default |
| Which self-hosted/cloud IdP(s) to validate the OIDC integration against first (e.g. Keycloak, Okta, Azure AD/Entra, Google Workspace) | Open — narrower than it used to sound: OIDC-only-at-MVP itself is already decided (architecture §3.4), this is just "which IdP do we test against" | Engineering | The OIDC client implementation ticket needs a concrete target to build and test against |
| Backoff/retry handling for GitLab's rate-limit responses (429 / `Retry-After`) in the ingestion scheduler | Narrower than previously stated — the two source docs disagreed with each other here. Tech-stack §5 implied this matters even at Phase 0 (GitLab is the Phase 0 host); architecture §6 said it wasn't needed until Phase 1's multi-host support. Resolution: basic 429/backoff handling is baseline correctness for *any* external API client and belongs in Phase 0's ingestion worker regardless of volume; what's genuinely Phase-1-only is the more elaborate *multi-host budget allocation* strategy, which only exists once there's more than one host to allocate a budget across. | Engineering | Nothing blocking — basic backoff is a normal part of the Phase 0 ingestion ticket, not a separate decision to make first |
| Whether/when to expose the doc-store bare repo(s) via a real git-over-HTTP endpoint, and what auth it would use | Deferred by design choice, not by indecision — the storage model (a real repo on disk) keeps this cheap to add later with no migration (tech-stack §4). Revisit only if there's an actual request for CLI-level access. | Product (revisit on demand) | Nothing at Phase 0 |
| Path to horizontal scaling of `app-backend` once ingestion needs to move off the in-process worker pool | Deferred — not a Phase 0/1 concern at this product's expected scale (architecture §6) | Engineering (revisit when ingestion volume demands it) | Nothing at Phase 0 |

## 2. Phase 2 features — already have a minimal starting design, not blocking anything

These aren't open questions anymore in the sense of "undecided" — each has a concrete MVP-appropriate design already written down. They're listed here only so Phase 2 ticketing has one place to start from, without re-reading every prior round of review.

| Feature | Design lives at | One-line summary |
|---|---|---|
| Contextual surfacing | overview §5.4/§6, architecture §6 | Admin-curated link table between a repo connection and doc-store pages — not automatic relatedness detection |
| Notifications/reminders | overview §6, architecture §6 | Email via the existing SMTP relay; Slack via a new bot-token/webhook credential (not yet added to deployment topology); recipient resolution goes through the `onboarding_hire` grant model, never an independently configured channel |
| Calendar integration | overview §6, architecture §6 | Server-generated `.ics` files only — no OAuth, no stored calendar credentials, no live sync (explicitly scoped down from full two-way sync) |
| Audit log | overview §6, architecture §6 | Postgres table, INSERT-only grants (no UPDATE/DELETE, including for admins), scoped to permission/repo-connection/auth events that git history doesn't already cover |
| Analytics | overview §6, architecture §6 | Aggregate queries over `started_at`/`completed_at` step timestamps (captured since Phase 0); "drop-off point" is well-defined because plan progression is gated |
| Finer-grained permissions | tech-stack §6, architecture §5 | `(user, role, resource_type, resource_id)` grants — schema exists since Phase 0 with a `global` sentinel; Phase 2 populates `doc_space`/`repo_connection`/`onboarding_hire` resource types against the same table |

## 3. Resolved — kept as a decision log

Full reasoning stays in the doc that resolved each one; not repeated here.

- Repo layout for the doc-store — one bare repo per company instance. See tech-stack §4, architecture §2.4.
- Conflict resolution for concurrent doc edits — optimistic locking, no merge UI at MVP. See architecture §2.4.
- Doc-store GC/repack cadence — daily `git gc --auto`. See architecture §2.4.
- Whether Next.js SSR is required — no, mostly client-rendered. See architecture §2.1.
- Which code host is Phase 0 — GitLab self-managed, not GitHub. See tech-stack §5.
- How much of doc versioning needs to be a real git repo — all of it; `go-git`-managed bare repo, no separate git server. See tech-stack §4.
- Self-hosted vs. hosted/managed deployment — self-hosted only for v1. See tech-stack §8.
- Permissions model shape at MVP — see §2 above (now itself a Phase 2 item with a Phase 0 schema foundation).
