# Sprint 5.2: The "Daily Focus" Algorithm

> **Epic:** 5 — The Reactive Command Center (Dashboard)
> **Depends On:** Sprint 5.1 (Agent Feed Aggregation), Sprint 1.1 (ContextState)
> **Objective:** Implement intelligent priority sorting for feed cards and apply traffic-light visual treatment.

---

## Sprint Tasks

### Task 5.2.1: Implement Priority Logic

**Status:** ⬜ Not Started

**Priority Tiers:**

| Priority | Trigger | Color | Example |
|----------|---------|-------|---------|
| P1 (Critical) | Safety incidents, blocking CPM tasks | 🔴 Red | "Electrical inspection failed — 3 tasks blocked" |
| P2 (Urgent) | Financial approvals pending > 48 hours | 🟡 Amber | "Invoice from Ace Plumbing awaiting approval (52 hrs)" |
| P3 (Routine) | Status updates, completions | 🟢 Green | "Framing phase complete ahead of schedule" |

**Current State:**
- [fb-home-feed.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/components/feed/fb-home-feed.ts) groups cards by horizon (`today`, `this_week`, `horizon`) but not by priority
- [FeedCard type](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/types/feed.ts) has `urgency` field
- [fb-feed-card.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/components/feed/fb-feed-card.ts) renders individual cards

**Atomic Steps:**

1. **Define priority scoring function:**
   ```ts
   // utils/feed-priority.ts [NEW]
   export type FeedPriority = 'critical' | 'urgent' | 'routine';
   
   export function scorePriority(card: FeedCard): { priority: FeedPriority; score: number } {
       // P1: Safety or CPM blocking
       if (card.type === 'safety_alert' || card.tags?.includes('blocking')) {
           return { priority: 'critical', score: 100 };
       }
       // P2: Pending approvals > 48 hours
       if (card.type === 'approval_pending') {
           const hours = hoursAgo(card.created_at);
           if (hours > 48) return { priority: 'urgent', score: 80 };
       }
       // P3: Everything else
       return { priority: 'routine', score: 20 };
   }
   ```

2. **Update `fb-home-feed._groupCards()`:**
   - Sort cards within each horizon group by priority score (descending)
   - P1 cards always appear first regardless of creation time

3. **Backend: Priority computation in FeedAggregator:**
   - When generating feed cards, compute priority from source data
   - CPM engine provides blocking task data
   - Financial service provides approval age
   - Include `priority` field in FeedCard API response

---

### Task 5.2.2: Visual Polish — Traffic Light Coloring

**Status:** ⬜ Not Started

**Atomic Steps:**

1. **Update `fb-feed-card.ts` styles:**
   ```css
   :host([priority="critical"]) .card {
       border-left: 4px solid #ef4444;
       background: linear-gradient(90deg, rgba(239,68,68,0.05) 0%, transparent 15%);
   }
   :host([priority="urgent"]) .card {
       border-left: 4px solid #f59e0b;
       background: linear-gradient(90deg, rgba(245,158,11,0.05) 0%, transparent 15%);
   }
   :host([priority="routine"]) .card {
       border-left: 4px solid #10b981;
   }
   ```

2. **Add priority icon/badge** to card header:
   - Critical: 🔴 pulsing dot + "CRITICAL" badge
   - Urgent: 🟡 dot + "NEEDS ATTENTION"
   - Routine: 🟢 subtle dot (no text)

3. **Add priority filter buttons** to feed header:
   - "All" | "Critical" | "Action Needed" | "Updates"
   - Clicking a filter shows only that priority tier

4. **Dashboard summary bar** at top of feed:
   - "3 Critical | 5 Action Needed | 12 Updates"
   - Clickable to jump to first card of that priority

---

## Codebase References

| File | Path | Lines | Notes |
|------|------|-------|-------|
| fb-home-feed.ts | `frontend/src/components/feed/fb-home-feed.ts` | 595 | Add priority sorting |
| fb-feed-card.ts | `frontend/src/components/feed/fb-feed-card.ts` | Existing | Add priority styling |
| fb-feed-section.ts | `frontend/src/components/feed/fb-feed-section.ts` | Existing | May need priority subgroups |
| feed.ts | `frontend/src/types/feed.ts` | Existing | Add `priority` to FeedCard type |
| feed-priority.ts | `frontend/src/utils/` | [NEW] | Priority scoring utility |

## Verification Plan

- **Manual:** View feed with mixed priority cards → verify critical cards appear first
- **Manual:** Verify traffic-light border colors match priority
- **Manual:** Click priority filter → verify only matching cards shown
- **Manual:** Verify P2 cards with pending approvals > 48hrs show amber treatment
- **Manual:** Verify dashboard summary bar counts are accurate
