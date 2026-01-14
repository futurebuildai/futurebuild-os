# Handoff: Phase 6 Step 43.6

**Previous Step:** 43.5 (Wiring & Assembly) - **COMPLETED** ✅
**Current Step:** 43.6 (Verification)

## Status
- **Step 43.5 Complete**: Wiring of `Orchestrator` and `ChatHandler` into `server.go` is done. Route `/api/v1/chat` is active and secured.
- **Ready for Step 43.6**: We need to verify the endpoint works as expected.

## Context for Step 43.6
This is the "Smoke Test" for the Chat Engine. We need to prove that:
1.  We can hit `POST /api/v1/chat`.
2.  Auth Middleware allows valid tokens and rejects others.
3.  The Orchestrator processes the message and saves it to the DB.
4.  We get a valid JSON response.

## Requirements
1.  **Create Integration Test**: `tests/integration/chat_test.go` (or similar).
2.  **Mock Auth**: Generate a valid JWT for the test user.
3.  **Execute Request**: Send a `POST /api/v1/chat` request.
4.  **Assert**:
    -   HTTP 200 OK.
    -   Response contains `reply` and `intent`.
    -   DB table `chat_messages` contains the user message and the model reply.

## Key Files
-   `test/integration/chat_test.go` (New)
-   `internal/server/server.go`

## Spec References
-   `PRODUCTION_PLAN.md` Step 43.6
