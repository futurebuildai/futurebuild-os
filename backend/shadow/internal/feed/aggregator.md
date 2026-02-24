# FeedAggregator Service — Shadow Spec

> **Package:** `internal/feed`
> **Sprint:** 5.1 — Agent Feed Aggregation
> **Status:** Shadow Spec (Frontend contract defined, Go implementation pending)

---

## Structs

```go
// FeedAggregator manages real-time fan-out of feed events to connected clients.
type FeedAggregator struct {
    subscribers map[string]chan FeedEvent  // keyed by session/user ID
    mu          sync.RWMutex
}

// FeedEvent represents a single mutation to a user's feed.
type FeedEvent struct {
    Type    string   `json:"type"`    // "card_added", "card_updated", "card_removed"
    Card    FeedCard `json:"card"`    // Full card for add/update; empty for remove
    CardID  string   `json:"card_id"` // Used for "card_removed" events
    Source  string   `json:"source"`  // "inspector", "gopher", "strategist", "system"
}
```

## Methods

```go
// NewFeedAggregator creates a new aggregator instance.
func NewFeedAggregator() *FeedAggregator

// Subscribe registers a client and returns a receive-only channel.
// The channel is buffered (cap 64) to absorb bursts.
func (a *FeedAggregator) Subscribe(userID string) <-chan FeedEvent

// Unsubscribe removes a client and closes its channel.
func (a *FeedAggregator) Unsubscribe(userID string)

// Publish sends an event to all subscribers. Non-blocking: drops if buffer full.
func (a *FeedAggregator) Publish(event FeedEvent)
```

## Interim Event Sources

Until agents exist, the aggregator is fed by:
1. **Invoice status changes** — when an invoice is approved/rejected, publish `card_added` or `card_updated`
2. **Schedule milestone completions** — when a task reaches 100%, publish `card_added`
3. **Manual admin broadcasts** — admin API endpoint to push arbitrary cards

## Frontend Contract

The frontend SSE client (`feed-sse.ts`) expects:
- Named SSE events: `card_added`, `card_updated`, `card_removed`
- `card_added` / `card_updated` data: JSON-serialized `FeedCard`
- `card_removed` data: JSON `{ "card_id": "<uuid>" }`

## FeedCard Reference

See `internal/models/feed_card.go` for the canonical `FeedCard` struct.
Frontend TypeScript mirror: `frontend/src/types/feed.ts` → `FeedCard` interface.
