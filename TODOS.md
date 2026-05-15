# TODOs

## Post-MVP: FSRS Parameter Fine-Tuning

**What:** Fine-tune all 17 FSRS parameters based on user feedback and deck statistics.

**Why:** MVP ships with a lossy Ease→Retention mapping and placeholder FSRS math. Once users report scheduling accuracy issues (cards feel too frequent or too rare), we'll have real data to optimize parameters.

**Impact:** Better scheduling = fewer wasted reviews = users learn faster.

**Context:** The Ease→Retention formula in the MVP is approximate (lossy). FSRS has 17 parameters that Helix's implementation uses. We're shipping with defaults; tuning comes post-MVP when we have user decks to validate against.

**Dependencies:** Blocks — Nothing. Post-MVP feature.

**Next steps:** Monitor user feedback on FSRS scheduling. After 1-2 releases, analyze deck statistics to identify which parameters need adjustment.

---

## Post-MVP: Per-Deck Algorithm Selection

**What:** Allow users to choose different algorithms for different decks (not app-wide only).

**Why:** Users have different learning styles per deck. Language learners might want FSRS (tight scheduling). Reference material might work better with SM-2 (familiar, predictable).

**Impact:** More flexible SRS experience, higher user satisfaction for power users.

**Context:** MVP is app-wide toggle (all decks use the same algorithm). Per-deck choice requires UI changes (settings per-deck, not global) and more complex data model (store algorithm on Card, not just Deck).

**Dependencies:** Post-MVP FSRS tuning should happen first (ship stable FSRS before letting users mix algorithms).

**Next steps:** Gather user feedback on app-wide toggle. If demand arises, design per-deck settings UI.

---
