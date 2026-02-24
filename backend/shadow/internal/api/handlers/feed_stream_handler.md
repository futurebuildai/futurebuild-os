# Feed Stream Handler — Shadow Spec

> **Package:** `internal/api/handlers`
> **Sprint:** 5.1 — Agent Feed Aggregation
> **Status:** Shadow Spec (Frontend contract defined, Go implementation pending)

---

## Endpoint

```
GET /api/v1/portfolio/feed/stream
```

**Auth:** Required (Bearer token). User ID extracted from JWT claims.

**Response:** `text/event-stream` (Server-Sent Events)

## Handler Signature

```go
// HandleFeedStream serves an SSE connection for real-time feed updates.
// It subscribes to the FeedAggregator and streams events until the client disconnects.
func HandleFeedStream(agg *feed.FeedAggregator) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        flusher, ok := w.(http.Flusher)
        if !ok {
            http.Error(w, "streaming unsupported", http.StatusInternalServerError)
            return
        }

        userID := middleware.GetUserID(r.Context())

        // Set SSE headers
        w.Header().Set("Content-Type", "text/event-stream")
        w.Header().Set("Cache-Control", "no-cache")
        w.Header().Set("Connection", "keep-alive")
        w.Header().Set("X-Accel-Buffering", "no") // nginx passthrough

        ch := agg.Subscribe(userID)
        defer agg.Unsubscribe(userID)

        for {
            select {
            case <-r.Context().Done():
                return
            case event, ok := <-ch:
                if !ok {
                    return
                }
                data, _ := json.Marshal(event.Card)
                if event.Type == "card_removed" {
                    data, _ = json.Marshal(map[string]string{"card_id": event.CardID})
                }
                fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, data)
                flusher.Flush()
            }
        }
    }
}
```

## SSE Wire Format

```
event: card_added
data: {"id":"...","org_id":"...","project_id":"...","card_type":"budget_alert",...}

event: card_updated
data: {"id":"...","priority":1,...}

event: card_removed
data: {"card_id":"uuid-here"}
```

## Keepalive

Send a comment line every 30 seconds to prevent proxy/LB timeout:
```
: keepalive
```

## Frontend Client

`frontend/src/services/feed-sse.ts` — connects to this endpoint.
Auto-reconnects with exponential backoff (2s base, 30s max).
