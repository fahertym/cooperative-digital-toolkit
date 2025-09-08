# 🗳️ Phase 2: Voting System Implementation Plan

## Status: ✅ Ready to Begin
- **Prerequisites Complete**: Proposals, Ledger, Announcements domains all working
- **All Tests Passing**: Comprehensive smoke tests verify end-to-end functionality  
- **Architecture Proven**: Modular domain pattern established across 3 domains

---

## 🎯 **Voting System Overview**

### **Goal**: Enable democratic decision-making on proposals with transparent, auditable voting

### **Core Requirements**:
- **Vote Casting**: Members vote for/against/abstain on open proposals
- **Quorum Management**: Configurable minimum participation thresholds
- **Real-time Tallies**: Live vote counts and outcome prediction
- **Audit Trail**: Immutable vote history for transparency
- **Auto-closure**: Proposals automatically close when quorum reached

### **Design Philosophy**: 
- Extend existing `proposals` domain rather than separate domain
- Maintain immutable audit trail for democratic accountability
- Support configurable quorum rules for different proposal types
- Ensure vote integrity with database constraints

---

## 🏗️ **Technical Architecture**

### **1. Data Model Extensions**

#### **Vote Table**
```sql
CREATE TABLE votes (
  id SERIAL PRIMARY KEY,
  proposal_id INTEGER NOT NULL REFERENCES proposals(id) ON DELETE CASCADE,
  member_id INTEGER NOT NULL,  -- Future: FK to members table
  choice TEXT NOT NULL CHECK (choice IN ('for', 'against', 'abstain')),
  notes TEXT,  -- Optional member reasoning
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(proposal_id, member_id)  -- One vote per member per proposal
);
```

#### **Vote Events Table (Audit Trail)**
```sql
CREATE TABLE vote_events (
  id SERIAL PRIMARY KEY,
  proposal_id INTEGER NOT NULL REFERENCES proposals(id) ON DELETE CASCADE,
  member_id INTEGER NOT NULL,
  event_type TEXT NOT NULL CHECK (event_type IN ('vote_cast', 'vote_changed', 'vote_retracted')),
  old_choice TEXT CHECK (old_choice IN ('for', 'against', 'abstain')),
  new_choice TEXT CHECK (new_choice IN ('for', 'against', 'abstain')),
  notes TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

#### **Quorum Configuration Table**
```sql
CREATE TABLE quorum_rules (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  minimum_participation_percent INTEGER NOT NULL CHECK (minimum_participation_percent > 0 AND minimum_participation_percent <= 100),
  minimum_for_percent INTEGER NOT NULL CHECK (minimum_for_percent > 0 AND minimum_for_percent <= 100),
  is_default BOOLEAN DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### **2. New Go Models**

```go
type Vote struct {
    ID         int32     `json:"id"`
    ProposalID int32     `json:"proposal_id"`
    MemberID   int32     `json:"member_id"`
    Choice     string    `json:"choice"`    // 'for', 'against', 'abstain'
    Notes      string    `json:"notes"`
    CreatedAt  time.Time `json:"created_at"`
}

type VoteEvent struct {
    ID         int32     `json:"id"`
    ProposalID int32     `json:"proposal_id"`
    MemberID   int32     `json:"member_id"`
    EventType  string    `json:"event_type"`  // 'vote_cast', 'vote_changed', 'vote_retracted'
    OldChoice  *string   `json:"old_choice,omitempty"`
    NewChoice  *string   `json:"new_choice,omitempty"`
    Notes      string    `json:"notes"`
    CreatedAt  time.Time `json:"created_at"`
}

type VoteTally struct {
    ProposalID           int32  `json:"proposal_id"`
    Status               string `json:"status"`           // proposal status: 'open', 'closed'
    TotalEligibleMembers int    `json:"total_eligible"`   // Future: from members table
    VotesCast            int    `json:"votes_cast"`
    ForVotes             int    `json:"for_votes"`
    AgainstVotes         int    `json:"against_votes"`
    AbstainVotes         int    `json:"abstain_votes"`
    QuorumMet            bool   `json:"quorum_met"`
    ParticipationPercent float64 `json:"participation_percent"`
    ForPercent           float64 `json:"for_percent"`
    Outcome              string  `json:"outcome"`         // 'passed', 'failed', 'pending'
    QuorumRule           string  `json:"quorum_rule"`     // name of applied rule
}

type QuorumRule struct {
    ID                        int32     `json:"id"`
    Name                      string    `json:"name"`
    MinimumParticipationPercent int      `json:"minimum_participation_percent"`
    MinimumForPercent         int       `json:"minimum_for_percent"`
    IsDefault                 bool      `json:"is_default"`
    CreatedAt                 time.Time `json:"created_at"`
}
```

### **3. API Endpoints (Extend Proposals Domain)**

```
POST /api/proposals/{id}/votes     - Cast or update vote
GET  /api/proposals/{id}/votes     - List all votes (with permissions)
GET  /api/proposals/{id}/tally     - Get current vote tally and outcome
DELETE /api/proposals/{id}/votes/{member_id} - Retract vote (if allowed)
POST /api/proposals/{id}/finalize  - Admin: Lock voting and finalize outcome

GET  /api/quorum-rules             - List available quorum rules
POST /api/quorum-rules             - Create new quorum rule (admin)
GET  /api/quorum-rules/{id}        - Get specific quorum rule
```

---

## 🔧 **Implementation Steps**

### **Step 1: Database Schema & Migrations**
- [ ] Create `0003_voting.sql` migration in proposals domain
- [ ] Add votes table with proper constraints and indexes
- [ ] Add vote_events table for immutable audit trail
- [ ] Add quorum_rules table with default configurations
- [ ] Test migration rollback scenarios

### **Step 2: Repository Layer Extensions** 
- [ ] Extend `proposals/repo.go` with voting methods
- [ ] Implement `CastVote(ctx, proposalID, memberID, choice, notes) error`
- [ ] Implement `GetVotes(ctx, proposalID) ([]Vote, error)`  
- [ ] Implement `GetVoteTally(ctx, proposalID) (VoteTally, error)`
- [ ] Implement `LogVoteEvent(ctx, event VoteEvent) error`
- [ ] Add quorum rule CRUD operations

### **Step 3: Business Logic & Validation**
- [ ] Vote validation: proposal must be open, one vote per member
- [ ] Quorum calculation logic with configurable rules
- [ ] Outcome determination: passed/failed/pending based on thresholds
- [ ] Auto-finalization when quorum definitively reached
- [ ] Vote change handling with proper audit trail

### **Step 4: HTTP Handlers**
- [ ] Extend `proposals/http.go` with voting endpoints
- [ ] Vote casting with JSON validation and member verification
- [ ] Tally endpoint with real-time calculation  
- [ ] Vote listing with proper permission checks
- [ ] Admin finalization endpoint with outcome locking

### **Step 5: Comprehensive Testing**
- [ ] Unit tests for all voting repository operations
- [ ] Integration tests for quorum calculation edge cases
- [ ] Audit trail verification tests
- [ ] Vote integrity constraint testing
- [ ] Performance testing with large member counts

### **Step 6: Integration & Documentation**
- [ ] Update smoke tests with voting workflow
- [ ] Update API documentation with voting endpoints
- [ ] Add voting examples to development guides
- [ ] Document quorum rule configuration options

---

## 🎯 **Key Implementation Decisions**

### **1. Single Domain Extension vs. New Domain**
**Decision**: Extend `proposals` domain
**Rationale**: Votes are tightly coupled to proposals; single domain reduces complexity

### **2. Vote Immutability**  
**Decision**: Allow vote changes with audit trail
**Rationale**: Members should be able to change their minds; audit trail preserves transparency

### **3. Quorum Rule Storage**
**Decision**: Database-driven configurable rules
**Rationale**: Different cooperatives need different thresholds; avoid hardcoded values

### **4. Auto-finalization**
**Decision**: Automatic closure when outcome is definitive
**Rationale**: Reduces administrative overhead; members see immediate results

### **5. Member Validation**
**Decision**: Use existing member_id references, validate in application layer
**Rationale**: Prepare for future auth integration without breaking current APIs

---

## 📊 **Expected Testing Scenarios**

### **Happy Path Tests**
- Create proposal → Cast votes → Meet quorum → Auto-close with outcome
- Vote changes with proper audit trail
- Tally calculation with various quorum rules
- CSV export with voting data included

### **Edge Case Tests**  
- Voting on closed proposals (should fail)
- Double voting attempts (should update, not duplicate)
- Quorum exactly at threshold boundaries
- Large vote counts (performance testing)
- Concurrent vote casting (race conditions)

### **Security Tests**
- Vote integrity constraints 
- Audit trail immutability
- Permission validation for vote viewing
- Proposal state consistency during voting

---

## 🚀 **Success Criteria**

### **Functional Requirements**
- [ ] Members can cast votes on open proposals
- [ ] Vote tallies update in real-time  
- [ ] Quorum rules are configurable and enforced
- [ ] Proposals auto-close when outcomes are definitive
- [ ] Complete audit trail of all voting activity
- [ ] CSV exports include voting data

### **Technical Requirements**
- [ ] All voting operations under 200ms response time
- [ ] Database constraints prevent vote integrity issues
- [ ] 100% test coverage for voting logic
- [ ] Backward compatibility with existing proposals API
- [ ] Smoke tests cover complete voting workflows

### **Democratic Requirements**
- [ ] Transparent vote tallies visible to all members
- [ ] Immutable audit trail for accountability
- [ ] Configurable quorum rules for cooperative autonomy
- [ ] Member vote privacy (notes optional, choices may be anonymous)

---

## 🔄 **Future Extensions (Post-MVP)**

### **Advanced Voting Features**
- **Ranked Choice Voting**: Multiple candidate proposals
- **Delegated Voting**: Proxy voting for absent members  
- **Time-limited Voting**: Automatic closure after deadline
- **Anonymous Voting**: Optional secret ballot mode
- **Vote Notifications**: Member alerts for new proposals requiring votes

### **Integration Enhancements**
- **Member Authentication**: Replace member_id with authenticated sessions
- **Email Notifications**: Vote reminders and outcome announcements
- **Mobile Push**: Real-time voting alerts
- **Calendar Integration**: Proposal deadlines and voting periods
- **External Audit**: Export voting records for regulatory compliance

---

## 📋 **Implementation Checklist**

### **Phase 2.1: Foundation (Week 1)**
- [ ] Design vote data models and constraints
- [ ] Create database migrations with proper indexes
- [ ] Implement basic vote casting in repository layer
- [ ] Add vote tally calculation logic

### **Phase 2.2: Core Features (Week 2)**
- [ ] Build HTTP endpoints for voting operations
- [ ] Implement quorum rule configuration
- [ ] Add audit trail logging for all vote events  
- [ ] Create comprehensive unit tests

### **Phase 2.3: Integration (Week 3)**
- [ ] Add voting to smoke tests and CI pipeline
- [ ] Update API documentation with voting specifications
- [ ] Test edge cases and performance scenarios
- [ ] Integrate with existing CSV export functionality

### **Phase 2.4: Polish (Week 4)**
- [ ] Add real-time tally updates
- [ ] Implement auto-finalization logic
- [ ] Create admin interfaces for quorum management
- [ ] Document cooperative best practices for voting

---

**Next Action**: Begin with Step 1 (Database Schema & Migrations) to establish the foundation for democratic voting in the Cooperative Digital Toolkit.

This voting system will complete the core governance functionality, enabling cooperatives to make democratic decisions with full transparency and accountability.
