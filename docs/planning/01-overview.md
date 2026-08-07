# cold_start — Planning Overview

> Preliminary project name: **cold_start**. Still open to changing later if a more product-facing name emerges — this is what the repository and codebase currently use.

## 1. Problem Statement

Developer onboarding at most companies is either absent or badly maintained: wikis go stale, "who do I ask about X" is tribal knowledge, and there's no single place that connects the *code* (who actually knows this service, is the README still true) with the *non-code* context (how does this company work, who's who, what's the process for X) a new hire needs.

This project is a self-hosted, web-based onboarding platform that:

1. Connects to existing code repositories and derives **meta information** from them (contribution patterns, likely subject-matter experts, documentation freshness) instead of asking humans to maintain a second copy of that knowledge.
2. Provides a **managed, rights-scoped space for non-code documentation** (company info, processes, team structure) that is version-controlled like code but editable by non-technical staff through a normal web UI — no git CLI, no full repo-hosting account required.
3. Wraps both in a **guided onboarding experience**: definable step-by-step plans per role/team, progress tracking, and surfacing the right document or contact at the right time.

## 2. Goals

- **G1 — Reduce time-to-productivity** for new engineering hires by centralizing what they need and removing stale/missing docs as a blocker.
- **G2 — Make expertise discoverable.** Answer "who can I ask about this code?" automatically from repo activity instead of relying on a person remembering.
- **G3 — Make documentation upkeep visible.** Surface docs that are likely stale (e.g., untouched while the code they describe changed heavily) rather than silently trusting them.
- **G4 — Let non-engineers own non-code content** (company handbook, policies, team info) without needing a developer-grade tool or a full seat on the main code host.
- **G5 — Make onboarding itself a first-class, configurable workflow**, not a one-off checklist in a doc nobody updates.

## 3. Non-Goals (at least for v1)

- Not a replacement for the company's primary code host (GitHub/GitLab/etc.) — it *connects to* it, doesn't replace it.
- Not a general-purpose wiki or CMS competitor — non-code doc support is scoped to what onboarding and team-context needs, not arbitrary publishing.
- Not an HRIS / payroll / benefits system. It may *link to* those, not replace them.
- Not initially multi-tenant SaaS — assume one instance per company to start; multi-tenancy is a possible later phase.

## 4. Target Users / Personas

| Persona | Needs |
|---|---|
| **New hire** | A guided plan of onboarding steps, easy access to relevant docs, and a way to find "who do I ask" without interrupting the whole team. |
| **Onboarding buddy / manager** | Ability to define and assign onboarding plans per role/team, and to see a new hire's progress. |
| **Engineer / tech lead** | Wants accurate auto-derived expertise mapping and doc-staleness signals without manual upkeep. Wants repo connections to be low-friction and read-only/safe by default. |
| **Non-technical staff (HR, ops, People team)** | Needs to write and update company/process documentation without learning git or needing a developer account. |
| **Admin** | Manages repo connections, rights/roles, and instance-wide configuration. |

## 5. Feature Pillars

### 5.1 Repository Intelligence
- Connect read-only to code repositories (GitHub/GitLab/Gitea/self-hosted, etc.).
- Derive **"who to ask"** suggestions per file/directory/service from commit history, blame, and review/PR activity — not a manually maintained CODEOWNERS-style list (though it can be seeded from one).
- Surface **documentation-freshness signals**: e.g., README/doc last touched vs. code churn in the same path, flag likely-stale docs.
- Repo-level dashboard: activity summary, key contacts, detected doc set, health signals.

### 5.2 Non-Code Documentation Management
- Version-controlled storage for non-code docs (company info, processes, glossaries, team pages), independent of the main code host's permission model.
- Web-based WYSIWYG/Markdown editing for users without git experience, gated by role-based permissions.
- Full history/diff/rollback, since it's git-backed underneath.
- Structured content types where useful (e.g., "team page," "process," "FAQ") vs. free-form pages.

### 5.3 Onboarding Planner
- Definable onboarding plans/templates (by role, team, or seniority) composed of **gated, ordered steps** — step N+1 unlocks only once step N is marked complete. This is a deliberate resolution (not the default "just a display order" reading): it makes "where do hires get stuck" a well-defined, single position per hire rather than an ambiguous set of incomplete steps in no particular relationship to each other, which the Phase 2 analytics feature (§6) depends on.
- Steps can reference: a doc page, a repo/README, a task ("get access to X"), a person to meet, or an external link.
- Progress tracking per new hire, with each step recording *when* it was started and completed, not just its current status — visibility for managers/buddies. The timestamps are captured from Phase 0 even though the analytics dashboard that reads them is Phase 2, since retroactively backfilling timing data for hires onboarded before that point isn't possible.
- Reusable templates so plans don't have to be rebuilt per hire.

### 5.4 Cross-Cutting: Search & Surfacing
- Unified search across connected repos' docs and internally managed docs.
- Contextual surfacing: e.g., when viewing a repo, show related non-code docs (team page, on-call process) and vice versa. Phase 2 (§6) — starting as admin-curated links between a repo connection and specific doc pages, not automatic relatedness detection.

## 6. Suggested Phasing

**Phase 0 — MVP**
- Single code-host connection (GitLab self-managed — see tech-stack §5) read-only ingestion of commits/blame for one or more repos.
- Basic "who to ask" per-path view.
- Git-backed non-code doc store with web editor and basic role permissions (viewer / editor / admin).
- Manually authored onboarding plan templates + per-hire progress checklist.

**Phase 1 — Freshness & Multi-source**
- Doc-staleness detection heuristics.
- Support additional code hosts / self-hosted git (GitHub.com, generic self-hosted git).
- Search across both doc sources.

**Phase 2 — Comfort & Scale**
- Notifications/reminders: email (reusing the SMTP relay already required for magic-link auth, tech-stack §6) and Slack (a new integration — bot token or incoming webhooks, not yet designed) for milestone reminders. Recipient selection must mirror the existing manager/buddy visibility scoping (5.3) rather than being configured independently (e.g., a Slack channel), so a reminder about a specific new hire's progress can't become a side-channel around the in-app RBAC model.
- Calendar integration for onboarding milestones: scoped to generating a downloadable `.ics` file — no OAuth, no live sync, no stored calendar credentials. Full two-way calendar sync (Google/Outlook OAuth) is explicitly out of scope for this phase.
- Finer-grained permission model: `(user, role, resource_type, resource_id)` grants, populating `doc_space`, `repo_connection`, and `onboarding_hire` resource types against the schema already built for it since Phase 0 (tech-stack §6). Manager/buddy onboarding visibility (5.3) is the first concrete instance of this, not a separate mechanism.
- Audit log: append-only by design (the writing path gets INSERT-only database grants, no UPDATE/DELETE — not even for admin-scoped app users, per standard audit-log practice) and scoped to what the doc-store's git history doesn't already cover: permission/role grant changes, repo-connection credential changes, and login/auth events — including the account-linking and session-invalidation events already defined in tech-stack §6. Deliberately excludes doc-store content edits, which already have full history via git (architecture §2.4).
- Analytics: onboarding completion time (per-hire elapsed time from plan assignment to final step completion, computed from the `started_at`/`completed_at` timestamps captured since Phase 0, §5.3) and drop-off points (aggregated across hires: the furthest gated step reached before a plan stalls — well-defined because progression is gated, §5.3). Individual-level timing is visible to a hire's manager/buddy through the `onboarding_hire` grant issued at plan assignment (architecture §3.3) — no new access-control surface beyond that grant; aggregate/company-wide views fold into the same retention/anonymization review already scoped for ingested commit-author data (tracked in [04-open-items.md](04-open-items.md)), since onboarding-completion behavior is comparably personal data about a specific employee.
- Contextual surfacing (5.4): admin-curated links between a repo connection and related non-code docs (e.g., a team page, an on-call process) — deliberately not automatic/heuristic matching at this stage, since that's a much larger and vaguer undertaking than the feature needs to start with.

## 7. Success Metrics (candidates)

- Time from hire start to "onboarding plan complete."
- % of onboarding steps completed without escalation to a manager/buddy.
- Doc staleness flags resolved vs. ignored (signal of whether the feature is trusted/used).
- Adoption: % of new hires whose onboarding runs through the tool vs. ad hoc.

## 8. Open Questions

All four questions originally posed here have since been resolved by decisions made in [02-tech-stack.md](02-tech-stack.md) and [03-architecture.md](03-architecture.md). Kept as a decision log rather than deleted, since it's useful provenance for why each was decided the way it was.

- ~~Which code hosts must be supported at launch?~~ — **GitLab self-managed first**, not GitHub, revised after review: it's the standard self-hosted code host for the enterprise segment this product targets, and self-hosted GitLab on the same internal network as cold_start sidesteps the public-reachability problems (webhooks, TLS/ACME) that GitHub.com introduces for a genuinely self-hosted deployment. GitHub.com and generic self-hosted git follow behind it, Bitbucket stays out of scope. See tech-stack §5.
- ~~How much of "non-code doc versioning" needs to be a *real* git repo vs. an internal version history model?~~ — a real repo: an in-process `go-git`-managed bare repo, owned entirely by the app, no separate git server. See tech-stack §4.
- ~~Self-hosted only, or should hosted/managed deployment be considered later?~~ — self-hosted only for v1 (docker-compose as the default target), consistent with this doc's own §3 non-goal ruling out multi-tenant SaaS at v1. See tech-stack §8.
- ~~Depth of the permissions model needed at MVP?~~ — viewer/editor/admin roles, modeled from day one as `(user, role, scope)` tuples so finer-grained scoping in Phase 2 doesn't require a data-model migration. See tech-stack §6.

## 9. Related Documents

- [02-tech-stack.md](02-tech-stack.md) — technology stack recommendation.
- [03-architecture.md](03-architecture.md) — system architecture derived from the stack.
- [04-open-items.md](04-open-items.md) — every decision still needed or deliberately deferred, tracked centrally.
- [05-architecture-diagrams.md](05-architecture-diagrams.md) — component, deployment, and data-flow diagrams for the architecture doc.
- [06-development-plan.md](06-development-plan.md) — Phase 0/MVP broken into epics and implementation tickets.
- [07-phase-1-development-plan.md](07-phase-1-development-plan.md) — Phase 1 (freshness & multi-source) epics and tickets.
- [08-phase-2-development-plan.md](08-phase-2-development-plan.md) — Phase 2 (comfort & scale) epics and tickets.
