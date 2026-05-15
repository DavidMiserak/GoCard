# Implementation Roadmap: Algorithm Selection Feature

Branch: `develop`  
Status: Design Finalized (Ready for Implementation)  
Last Updated: 2026-05-15

---

## Overview

This document defines the implementation roadmap for the algorithm selection feature (SM-2 vs FSRS). The design has been revised to address 5 critical architectural blockers identified during engineering review.

**Key Documents:**
- Design: `/home/david/.gstack/projects/GoCard/david-develop-design-revised-20250515.md` (FINAL)
- Test Plan: `/home/david/.gstack/projects/GoCard/david-develop-eng-review-test-plan-revised-20250515.md` (FINAL)
- User Stories: Issue #78 (GitHub)

---

## Critical Blockers & Solutions

| Blocker | Solution | Impact |
|---------|----------|--------|
| Rollback without backup | Add `EaseBackup` field to Card | Enables exact Ease restoration from FSRS→SM-2 |
| Persistence location unclear | Store `Algorithm` in markdown frontmatter | Disk is source of truth, roundtrip-safe |
| Ease→Retention lossy conversion | Accept lossy forward; rely on backup for reverse | No data loss, recovery is lossless |
| Atomic toggle unspecified | In-memory backup + all-or-nothing writes | Graceful failure recovery, no partial state |
| FSRS timing unclear | Defer parameter tuning to post-MVP | MVP uses approximate formula, captured in TODOS.md |

---

## Data Model Changes

### Card (internal/model/card.go)
**Add:**
```go
Algorithm   string    // "SM2" or "FSRS"
Retention   float64   // FSRS primary parameter (0-1)
EaseBackup  float64   // Snapshot of Ease at toggle time for rollback
```

### Deck (internal/model/deck.go)
**Add:**
```go
Algorithm   string    // "SM2" (default) or "FSRS", app-wide choice
```

### FrontMatter (internal/data/markdown_parser.go)
**Add to YAML:**
```yaml
algorithm: "SM2" or "FSRS"
retention: <float64>          # Only if Algorithm=="FSRS"
ease_backup: <float64>        # Only if backup exists
```

**Backwards Compatibility:**
- Old files (no `algorithm` field) default to `Algorithm="SM2"` on read
- Old GoCard versions ignore unknown `algorithm` field (graceful degradation)

---

## Implementation Phases

### Phase 1: Data Model & Persistence (Est. 2 days)
**Goal:** All three fields added to Card/Deck/FrontMatter with full roundtrip support.

**Tasks:**
1. Add fields to Card struct
2. Add field to Deck struct
3. Update FrontMatter struct with new YAML fields
4. Update markdown parser:
   - Read `algorithm`, `retention`, `ease_backup` from YAML
   - Default `algorithm="SM2"` for old files
5. Update markdown writer:
   - Write new fields to YAML frontmatter
   - Omit `ease_backup` if it's 0 (or always write)
6. **Test:** Markdown roundtrip test (read/write with new fields)

**Success Criteria:**
- ✅ Card.Algorithm persists to disk and survives app restart
- ✅ Old markdown files read correctly (default to SM-2)
- ✅ New markdown files roundtrip without loss

---

### Phase 2: Scheduler Refactoring & Dispatch (Est. 2 days)
**Goal:** Extract SM-2 logic into SM2Scheduler, implement dispatcher, stub FSRSScheduler.

**Tasks:**
1. Create Algorithm interface:
   ```go
   type Algorithm interface {
       Schedule(card *Card, rating int) ScheduleResult
   }
   ```
2. Refactor existing ScheduleCard() into `SM2Scheduler.Schedule()`:
   - Extract switch statement on rating (case 1-5)
   - Keep all existing logic unchanged
   - Return ScheduleResult with updated Ease, Interval, NextReview
3. Implement `FSRSScheduler.Schedule()` stub:
   - Use approximate Helix formula (placeholder)
   - Return ScheduleResult with updated Retention, Interval, NextReview
   - **Note:** MVP formula is lossy; post-MVP tuning improves accuracy
4. Create dispatcher in `internal/srs/algorithm.go` or `dispatcher.go`:
   ```go
   func ScheduleCard(card *Card, rating int) (*Card, error) {
       switch card.Algorithm {
       case "SM2": return SM2Scheduler{}.Schedule(...)
       case "FSRS": return FSRSScheduler{}.Schedule(...)
       default: return nil, fmt.Errorf(...)
       }
   }
   ```
5. Update all callers (SaveCardReview, review loop) to call dispatcher
6. **Test:** Unit tests for both schedulers independently

**Success Criteria:**
- ✅ SM-2 review produces same results as before (no behavior change)
- ✅ FSRS review produces different intervals than SM-2 (visual proof)
- ✅ Dispatcher correctly routes to each scheduler

---

### Phase 3: Toggle Logic & Atomic Migration (Est. 3 days)
**Goal:** Implement SM-2→FSRS and FSRS→SM-2 toggling with backup/rollback.

**Tasks:**
1. Implement `ToggleAlgorithm()` function:
   - Load all cards from disk
   - Create in-memory backup (deep copy)
   - Convert all cards (apply formula or restore from backup)
   - Write all cards atomically
   - On write failure: restore from in-memory backup, return error
2. Forward path (SM-2→FSRS):
   ```go
   for card in deck.Cards:
     card.Retention = (card.Ease - 1.0) / 3.0  // lossy conversion
     card.EaseBackup = card.Ease               // snapshot for rollback
     card.Algorithm = "FSRS"
   ```
3. Reverse path (FSRS→SM-2):
   ```go
   for card in deck.Cards:
     card.Ease = card.EaseBackup  // exact restoration
     card.Retention = 0           // clear FSRS state
     card.EaseBackup = 0          // backup no longer needed
     card.Algorithm = "SM2"
   ```
4. Error handling:
   - Catch write errors (disk full, permission denied, etc.)
   - Restore all cards from in-memory backup on error
   - Return error to UI with clear message
5. **Test:** Integration tests for toggle on 100+ card deck, failure recovery

**Success Criteria:**
- ✅ Toggle SM-2→FSRS completes < 2 seconds on 1000+ card deck
- ✅ All cards convert atomically (no partial state)
- ✅ EaseBackup preserves exact Ease (bitwise equality)
- ✅ Rollback FSRS→SM-2 restores exact Ease from EaseBackup
- ✅ Write failure triggers rollback (all cards reverted)

---

### Phase 4: Settings UI (Est. 1 day)
**Goal:** Add Settings menu with algorithm toggle.

**Tasks:**
1. Add "Settings" option to main menu (alongside Study, Browse, Statistics, Quit)
2. Implement Settings screen:
   - Display current algorithm
   - Radio buttons or tabs for SM-2 / FSRS selection
   - Info text: "Switching algorithms will migrate all cards. Original data is preserved."
3. Implement toggle flow:
   - On "Confirm": call `ToggleAlgorithm()`
   - On error: show error message, keep current algorithm
   - On success: show confirmation, return to menu
4. **Test:** Manual QA on settings navigation

**Success Criteria:**
- ✅ Settings menu navigates without crashes
- ✅ Algorithm toggle UI displays current choice
- ✅ Confirmation dialog explains action
- ✅ Error messages are clear

---

### Phase 5: Testing & QA (Est. 2 days)
**Goal:** Comprehensive test coverage for toggle flow, formula validation, rollback guarantee.

**Tests to Write:**

**Unit Tests:**
- [ ] Ease→Retention formula within tolerance (5 test cases)
- [ ] EaseBackup exactness (100 cards, bitwise equality)
- [ ] SM2Scheduler for all ratings (1-5)
- [ ] FSRSScheduler for all ratings (1-5)
- [ ] Dispatcher routing
- [ ] Rollback exactness (no re-conversion)

**E2E Tests:**
- [ ] Forward toggle (SM-2→FSRS) on 50-card deck
- [ ] Reverse toggle (FSRS→SM-2) with rollback
- [ ] Review cards in FSRS, intervals differ from SM-2
- [ ] Markdown roundtrip (toggle, close, reopen)
- [ ] Large deck (1000+ cards) < 2s, all-or-nothing atomicity
- [ ] Write failure simulation (rollback recovery)

**Manual QA:**
- [ ] Test on real deck (100+ cards)
- [ ] Test on mixed card history (old + new)
- [ ] Settings UI navigation
- [ ] Backwards compatibility (old GoCard on new markdown)

**Timebox:** 2 days for implementation + review

---

## Deferred Scope (Post-MVP)

**Captured in TODOS.md:**

1. **FSRS Parameter Fine-Tuning**
   - MVP uses approximate Helix formula
   - Post-MVP: collect user feedback, tune 17 parameters
   - Timeline: after 1-2 releases with real user data

2. **Per-Deck Algorithm Selection**
   - MVP: app-wide toggle only (Deck.Algorithm)
   - Post-MVP: per-card override (Card.Algorithm can differ from Deck.Algorithm)
   - Requires: per-deck settings UI, more complex card model

---

## Risk Mitigation

| Risk | Mitigation |
|------|-----------|
| Data corruption during toggle | In-memory backup + all-or-nothing writes; rollback on error |
| Lossy Ease→Retention conversion | EaseBackup preserves exact Ease; rollback is lossless |
| Large deck (1000+) timeout | Target < 2s; benchmark on actual hardware; optimize if needed |
| Write failures (disk full, etc.) | Catch error, restore from backup, user can retry |
| Backwards compatibility break | Old GoCard ignores unknown `algorithm` field; graceful degradation |

---

## Branch & Release Plan

**Branch:** `develop` (current)  
**Target Release:** v0.4.0 (next semver after v0.3.0)  
**Release Timeline:** End of iteration (1-2 weeks after Phase 5 complete)

**Pre-Release Checklist:**
- [ ] All tests passing (unit + E2E + manual QA)
- [ ] Code review approved
- [ ] Documentation updated (README, CHANGELOG)
- [ ] No data loss or corruption reported in testing
- [ ] TODOS.md updated with post-MVP items

**Release Process:**
1. Merge develop → main
2. Tag v0.4.0
3. Update releases page on GitHub
4. Announce feature (Reddit, discussions, etc.)

---

## Key Metrics for Success

- **Atomicity:** 100% all-or-nothing (0 partial toggles)
- **Performance:** Toggle < 2 seconds on 1000-card deck
- **Data Integrity:** 0 data loss incidents (EaseBackup exact match)
- **User Experience:** Clear settings UI, helpful confirmation/error messages
- **Backwards Compatibility:** Old GoCard versions don't break on new markdown
- **Test Coverage:** ≥90% code coverage for algorithm logic

---

## Questions & Open Items

1. **FSRS Formula:** Use Helix v1.x or other implementation? *(Answer: Helix v1.x, deferred tuning to post-MVP)*
2. **EaseBackup Field:** Always write to markdown or only if non-zero? *(Answer: Write only if non-zero, to keep old markdown readable)*
3. **Default Retention for New Cards:** 0.5 or derive from Ease? *(Answer: 0.5 default for new FSRS cards)*
4. **Rollback UI:** Show warning before toggle? *(Answer: Yes, info text: "Original data preserved")*

---

## Summary

The design revision addresses all 5 blockers identified in engineering review:

1. ✅ **Backup/Rollback** — EaseBackup enables exact restoration
2. ✅ **Persistence** — Algorithm stored in markdown frontmatter
3. ✅ **Lossy Conversion** — Backup makes reverse path lossless
4. ✅ **Atomic Migration** — In-memory backup + all-or-nothing writes
5. ✅ **FSRS Timing** — Deferred to post-MVP, MVP uses approximate formula

The implementation is well-scoped, low-risk, and ready to begin. Estimated effort: **2 weeks** for complete MVP with full test coverage.
