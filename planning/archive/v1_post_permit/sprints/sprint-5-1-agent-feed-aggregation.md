# Sprint 5.1: Agent Feed Aggregation

> **Epic:** 5 — The Reactive Command Center (Dashboard)
> **Depends On:** Sprint 1.1 (ContextState for scope filtering)
> **Objective:** Create a backend FeedAggregator that listens to all Agents and pushes updates to the dashboard in real-time via SSE.

---

## Sprint Tasks

### Task 5.1.1: Backend — Create `FeedAggregator` Service

**Status:** ✅ Complete

**Concept:** A centralized service that subscribes to events from Inspector, Gopher, and Strategist agents, then produces unified feed cards.

**Current State:**
- No backend agent management layer exists (shadow stubs only)
- Frontend [feed-sse.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/services/feed-sse.ts) (155L) — SSE client exists and is functional
- Frontend [fb-home-feed.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/components/feed/fb-home-feed.ts) (595L) — feed rendering exists with horizon groups
- Backend currently serves feed via `GET /api/v1/portfolio/feed` (REST, polled)

**Required Implementation:**

1. **Create `backend/internal/feed/aggregator.go`** [NEW]:
   ```go
   type FeedAggregator struct {
       subscribers map[string]chan FeedEvent  // keyed by session/user ID
       mu          sync.RWMutex
   }
   
   type FeedEvent struct {
       Type    string   `json:"type"`    // "card_added", "card_updated", "card_removed"
       Card    FeedCard `json:"card"`
       Source  string   `json:"source"`  // "inspector", "gopher", "strategist"
   }
   
   func (a *FeedAggregator) Subscribe(userID string) <-chan FeedEvent
   func (a *FeedAggregator) Unsubscribe(userID string)
   func (a *FeedAggregator) Publish(event FeedEvent)
   ```

2. **Create `backend/internal/api/handlers/feed_stream_handler.go`** [NEW]:
   ```go
   // GET /api/v1/portfolio/feed/stream — SSE endpoint
   func HandleFeedStream(w http.ResponseWriter, r *http.Request) {
       flusher, ok := w.(http.Flusher)
       // Set SSE headers
       // Subscribe to FeedAggregator
       // Loop: write events, flush
   }
   ```

3. **Agent integration points** (wire as agents are built):
   - Inspector Agent → publishes safety/compliance alerts
   - Gopher Agent → publishes material procurement events
   - Strategist Agent → publishes schedule risk analysis

4. **Interim:** Until agents exist, feed the aggregator from:
   - Invoice status changes (approved/rejected)
   - Schedule milestone completions
   - Manual admin broadcasts

---

### Task 5.1.2: Frontend — Implement `FeedSSE` Push Updates

**Status:** ✅ Complete

**Current State:**
- [feed-sse.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/services/feed-sse.ts) — SSE client already exists with `connect()`, `subscribe()`, `disconnect()`, `_scheduleReconnect()`
- Handles `card_added`, `card_updated`, `card_removed` event types
- [fb-home-feed.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/components/feed/fb-home-feed.ts) — renders grouped feed cards

**Atomic Steps:**

1. **Connect SSE on auth:** In `fb-home-feed.ts` `connectedCallback()`, call `feedSSE.connect()`
2. **Subscribe to events:** Add handler that processes incoming feed events:
   ```ts
   feedSSE.subscribe((event) => {
       switch (event.type) {
           case 'card_added':
               this._cards = [event.card, ...this._cards];
               break;
           case 'card_updated':
               this._cards = this._cards.map(c => c.id === event.card.id ? event.card : c);
               break;
           case 'card_removed':
               this._cards = this._cards.filter(c => c.id !== event.card_id);
               break;
       }
   });
   ```
3. **Context filtering:** Apply `store.contextState$` filter — only show cards matching current scope
4. **Animation:** New cards should slide in with a subtle animation (CSS `@keyframes slideIn`)
5. **Disconnect on destroy:** In `disconnectedCallback()`, call `feedSSE.disconnect()`
6. **Connection status indicator:** Show subtle dot (green=connected, red=disconnected) in feed header

---

## Codebase References

| File | Path | Lines | Notes |
|------|------|-------|-------|
| feed-sse.ts | `frontend/src/services/feed-sse.ts` | 155 | SSE client (exists, functional) |
| fb-home-feed.ts | `frontend/src/components/feed/fb-home-feed.ts` | 595 | Feed rendering (needs SSE wiring) |
| fb-feed-card.ts | `frontend/src/components/feed/fb-feed-card.ts` | Existing | Individual card component |
| fb-feed-section.ts | `frontend/src/components/feed/fb-feed-section.ts` | Existing | Horizon section grouping |
| mock-feed-service.ts | `frontend/src/services/mock-feed-service.ts` | Existing | Mock data (keep as fallback) |
| feed.ts | `frontend/src/types/feed.ts` | Existing | FeedCard, FeedSSEEvent types |
| aggregator.go | `backend/internal/feed/` | [NEW] | Event aggregation service |
| feed_stream_handler.go | `backend/internal/api/handlers/` | [NEW] | SSE HTTP handler |

## Verification Plan

- **Manual:** Open dashboard → verify SSE connection established (check Network tab for EventSource)
- **Manual:** Trigger a backend event (approve invoice) → verify new card appears without page refresh
- **Manual:** Verify card slide-in animation
- **Manual:** Disconnect network → verify reconnection indicator and auto-reconnect
- **Automated:** Load test: Push 100 events in 10 seconds → verify no dropped events on frontend
