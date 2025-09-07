# Cooperative Digital Toolkit — Product Requirements Document (PRD)

## 1. Vision & Mission
**Vision:** A member-owned digital commons where cooperatives of all sizes can run governance, finance, and collaboration on shared infrastructure aligned with cooperative values.

**Mission:** Deliver a modular, accessible, FOSS-first toolkit that reduces operational burden, increases democratic participation, and bridges sectors (worker, housing, food, ag, credit unions) across New York and beyond.

**Why now:** NY’s co-op economy is vibrant but fragmented. Downstate has dense worker/housing activity; upstate has foundational ag, utility, and credit union co-ops. Common gaps: governance friction, duplicated tooling, compliance overhead, and limited technical capacity. A shared stack lowers cost and raises capability.

## 2. Goals & Non-Goals
**Goals**
- MVP usable by real pilots within 6 months.
- 30-day “Member-Benefit Test”: a small co-op should reach first vote + first ledger entry inside 30 days of kickoff.
- Offline-first UX; low bandwidth friendly.
- Transparent logs for audits and democratic accountability.
- Path to federation (multiple co-ops, shared infra, portable identity).

**Non-Goals (for MVP)**
- Full accounting system. We integrate with QuickBooks/Xero/Odoo instead.
- POS replacement. We bridge minimal flows for food co-ops.
- Complex multi-tenant billing. Start simple; evolve after pilots.

## 3. Target Users & Personas
- **Member**: participates in proposals and voting; reads announcements.
- **Admin/Board**: configures roles, creates proposals, exports reports.
- **Federation steward**: coordinates multiple co-ops, audits activity.
- **Developer/Integrator**: sets up hosting; builds automations/integrations.

## 4. MVP Scope
1) **Governance & Membership**
   - Member directory + role-based permissions (RBAC)
   - Proposals, quorum rules, async voting
   - Meetings + minutes (basic)
2) **Finance (light)**
   - Simple ledger: dues, contributions, patronage accrual placeholders
   - Export to existing accounting via n8n (CSV/API)
3) **Portal & Comms**
   - Announcements, basic tasks
   - Audit log spanning modules
4) **Resilience**
   - Offline-first cache; conflict resolution strategy
   - CORS + auth foundations

## 5. Out-of-Scope (MVP)
- Real-time chat
- Inventory/POS engines
- Complex patronage distributions and tax filings (beyond placeholders)

## 6. Success Metrics (12 months)
- Adoption: 6–10 co-ops onboarded across sectors.
- Time-to-first-vote: ≤30 days from kickoff for 80% of pilots.
- Admin hours saved: ≥15–20% vs. baseline on governance + bookkeeping tasks.
- Participation: ≥70% members active on portal during a vote.

## 7. Pilots
- **Upstate conversions**: retiring businesses converting to co-ops (governance flows + ledger exports).
- **Downstate food co-ops**: member IDs, patronage accrual placeholders, minimal POS bridge.
- **Urban housing**: notices, maintenance intake, votes, compliance audit trail.

## 8. Architecture (see 20-architecture-overview.md)
- Phase 1: **Modular monolith** (Go backend, Svelte frontend, PostgreSQL). Event-logged actions for audits.
- Phase 2: selective service extraction if metrics justify; CockroachDB path for federation.
- Integrations via **n8n** and **Airbyte** for imports.
- Identity: WebAuthn now; DID/VC later.

## 9. Requirements & Acceptance
**Governance**
- Create proposal with title/body → members can view → vote window set → tally with quorum rule → immutable event recorded.
- Admin can export proposal results and minutes as CSV/PDF.

**Ledger (light)**
- Record cash-in/cash-out entries with type and notes → export CSV compatible with QuickBooks/Xero → audit log event.

**Portal**
- Announce message → visible to members → marked read/unread → event logged.

**Resilience**
- Browser offline: view cached proposals/announcements; queue create actions; sync when online.

## 10. Constraints
- FOSS license (Apache-2.0), self-hostable, low TCO.
- Green ops, small binaries, minimal RAM.
- Accessibility: keyboard navigable; WCAG-aware components.

## 11. Risks & Mitigations
- **Capital/hosting know-how**: provide one-click Docker + docs; offer hosted option via partner co-op.
- **Fragmentation**: publish APIs and RFCs; steering committee across sectors.
- **Lock-in pressure**: permissive license, documented exports, DID portability.

## 12. Roadmap
- M0–M1: MVP vertical slice (proposal create/list/vote; simple ledger; portal).
- M2–M4: Pilot onboarding and iteration; exports; training.
- M5–M6: Identity hardening, report packs, federation plan; publish adopters’ guide.

## 13. Open Questions
- Best first POS surfaces for food co-ops bridging?
- Which identity claims are most useful for federation (e.g., “member in good standing”)?
- Data commons governance model for anonymous benchmarks.

