# 🚀 Next Steps Implementation Plan

## Current Status: ✅ Ledger Domain Complete

The Cooperative Digital Toolkit now has:
- ✅ **Proposals API** - Complete governance workflow (open→closed, CSV export)
- ✅ **Ledger API** - Financial tracking with QuickBooks/Xero CSV export
- ✅ **Project Infrastructure** - Modular domain pattern, comprehensive testing, smoke tests

## 🎯 Phase 1: Announcements Domain (Immediate Priority)

### Overview
Implement member communications system with read/unread status tracking, priority levels, and activity feeds.

### Implementation Tasks

#### 1. **Data Models & Migrations** 
```sql
-- announcements table
CREATE TABLE announcements (
  id SERIAL PRIMARY KEY,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  author_id INTEGER, -- Future: FK to members table  
  priority TEXT DEFAULT 'normal', -- low, normal, high, urgent
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);

-- announcement_reads junction table  
CREATE TABLE announcement_reads (
  announcement_id INTEGER REFERENCES announcements(id),
  member_id INTEGER, -- Future: FK to members table
  read_at TIMESTAMPTZ DEFAULT now(),
  PRIMARY KEY (announcement_id, member_id)
);
```

#### 2. **Go Domain Structure**
Following established patterns from `proposals/` and `ledger/`:
- `backend/internal/announcements/model.go` - Announcement + AnnouncementRead structs
- `backend/internal/announcements/repo.go` - CRUD operations with read status joins
- `backend/internal/announcements/http.go` - HTTP handlers with member context
- `backend/internal/announcements/routes.go` - Chi router mounting
- `backend/internal/announcements/http_test.go` - Comprehensive unit tests
- `backend/internal/announcements/migrations/` - SQL files with constraints

#### 3. **API Endpoints**
```
GET  /api/announcements          - List with read status per member
POST /api/announcements          - Create new announcement (admin only)
GET  /api/announcements/{id}     - Get specific announcement
POST /api/announcements/{id}/read - Mark as read for authenticated member
GET  /api/announcements/unread   - Count unread announcements
```

#### 4. **Key Features**
- **Priority System**: Visual indicators for urgent announcements
- **Read Status**: Per-member tracking with timestamps
- **Filtering**: By priority, read status, date ranges
- **Pagination**: Support for large announcement volumes
- **Markdown Support**: Rich text formatting in announcement body

### Expected Deliverables
- Complete announcements domain following established patterns
- Unit tests achieving >90% coverage
- Updated smoke tests exercising all endpoints  
- API documentation in `/docs/22-api-spec.md`
- Updated data models in `/docs/23-data-models.md`

---

## 🎯 Phase 2: Voting System (Democracy Tools)

### Overview
Enable democratic decision-making on proposals with quorum rules, vote tallying, and immutable audit logs.

### Core Features
- **Vote Casting**: for/against/abstain on open proposals
- **Quorum Management**: Configurable minimum participation thresholds  
- **Real-time Tallies**: Live vote counts and outcome prediction
- **Immutable Audit**: Complete vote history for transparency
- **Result Automation**: Auto-close proposals when quorum met

### API Extensions
```
POST /api/proposals/{id}/votes   - Cast vote (authenticated members)
GET  /api/proposals/{id}/votes   - List votes (admin during, public after)
GET  /api/proposals/{id}/tally   - Get current vote tally and quorum status  
POST /api/proposals/{id}/finalize - Close voting and lock results (admin)
```

### Implementation Strategy
- Extend existing `proposals` domain rather than separate domain
- Add `votes` and `vote_events` tables for audit trail
- Implement quorum rules as configurable business logic
- Ensure vote integrity with database constraints

---

## 🎯 Phase 3: Authentication System (Member Identity)

### Overview  
Replace dev "Admin mode" toggle with production authentication supporting WebAuthn and email magic links.

### Authentication Options
1. **WebAuthn (Preferred)**: Passwordless biometric/security key auth
2. **Email Magic Links**: Secure tokenized login via email
3. **Hybrid Approach**: WebAuthn primary, email fallback

### Core Components
- Member registration and profile management
- Session handling with server-side storage
- RBAC: roles (admin, member, guest) with permissions
- CSRF protection and security headers

### Integration Impact
- All existing APIs will require member context
- Announcements read status tied to authenticated members
- Voting restricted to eligible members only
- Admin operations (create announcements, close proposals) role-gated

---

## 🎯 Phase 4: Frontend Migration & UX

### Overview
Migrate from basic Svelte to SvelteKit with offline-first architecture and accessibility.

### Technical Migration
- **SvelteKit Setup**: SSR, routing, form actions
- **Offline Support**: IndexedDB cache, queue failed requests
- **PWA Features**: Installable, background sync
- **Accessibility**: WCAG 2.1 AA compliance, screen reader support

### User Experience Priorities
- **Mobile-first**: Responsive design for co-op members
- **Low-bandwidth**: Minimal JS, progressive enhancement  
- **Cooperative Values**: Clear, democratic interfaces
- **30-day Onboarding**: Guided setup for new co-ops

---

## 📋 Development Priorities (Next 2-3 Sprints)

### Sprint 1: Announcements Foundation
- [ ] Create announcements domain structure
- [ ] Implement database models and migrations  
- [ ] Build repository layer with read status joins
- [ ] Create HTTP handlers and routes
- [ ] Write comprehensive unit tests
- [ ] Update smoke tests and API documentation

### Sprint 2: Announcements Polish  
- [ ] Add priority filtering and visual indicators
- [ ] Implement pagination for large datasets
- [ ] Add markdown support for rich announcements
- [ ] Create member activity feed views
- [ ] Performance optimization and caching

### Sprint 3: Voting System Foundation
- [ ] Design vote data models and audit tables
- [ ] Implement vote casting and validation logic
- [ ] Build quorum calculation and threshold management
- [ ] Create tally endpoints with real-time updates
- [ ] Add immutable vote event logging

---

## 🎯 Success Metrics

### Technical Metrics
- **API Coverage**: All domains have >90% test coverage
- **Performance**: <200ms response times for common operations
- **Reliability**: 99.9% uptime during pilot deployments
- **Security**: Zero critical vulnerabilities in security audit

### Product Metrics  
- **30-day Adoption**: Co-ops reach first vote within 30 days
- **Member Engagement**: >80% members read announcements within 48h
- **Democratic Participation**: >75% eligible members vote on proposals
- **Financial Transparency**: Monthly ledger exports to accounting systems

---

## 🔧 Developer Experience

### Current DX Strengths
- ✅ Modular domain pattern established
- ✅ Comprehensive testing with smoke tests
- ✅ Auto-migrating database schema  
- ✅ Clear API documentation and examples
- ✅ Consistent error handling and responses

### DX Improvements Planned
- **Hot Reload**: Frontend development server integration
- **API Explorer**: Interactive documentation with live examples  
- **Development Data**: Seed scripts for realistic test data
- **Docker Compose**: One-command local development environment
- **CI/CD**: Automated testing and deployment pipelines

---

## 🤝 Cooperative Values Integration

Every technical decision aligns with cooperative principles:

- **Democratic Control**: Transparent voting, member-owned data
- **Member Benefit**: User-centric design, accessible interfaces  
- **Community**: Open source, federation-ready architecture
- **Sustainability**: Low TCO, minimal vendor dependencies
- **Education**: Clear documentation, onboarding materials
- **Caring for Others**: Inclusive design, accessibility-first

---

*This plan serves as the roadmap for building production-ready cooperative digital tools that empower democratic governance and transparent financial management.*
