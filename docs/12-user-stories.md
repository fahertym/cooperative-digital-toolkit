# Epics & User Stories

## Epic: Proposals & Voting (MVP)
**Goal:** Members can propose, discuss (basic), and vote asynchronously.

- As an *admin*, I can create a proposal (title/body) so members can vote.
  - Given a valid title, when I submit, then a proposal is created and visible in list.
- As a *member*, I can read proposal details so I can make an informed decision.
- As an *admin*, I can set a vote window and quorum (Phase 2).
- As a *member*, I can cast a vote and see confirmation (Phase 2).
- As an *admin*, I can export results and minutes.

- As an *admin*, I can close a proposal so no further changes can occur.
  - When I call `POST /api/proposals/{id}/close` on an open proposal, then it returns 200 and status changes to `closed`; further close attempts return 409.

## Epic: Simple Ledger (MVP-light)
- As an *admin*, I can record dues/contributions with notes.
- As an *admin*, I can export ledger entries to CSV for QuickBooks/Xero.

## Epic: Portal & Announcements (MVP)
- As an *admin*, I can post an announcement visible to all members.
- As a *member*, I can mark announcements read/unread.

## Non-Functional
- Offline read; queued create; sync conflict policy documented.
- Audit trail for create/update actions.

# Epics & User Stories

Template:
```
As a <role>, I want <capability> so that <outcome>.
Acceptance:
- Given/When/Then ...
```
