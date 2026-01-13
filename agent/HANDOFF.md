# Handoff: Phase 6 Step 43.2

**Previous Step:** 43.1 (Chat Domain Modeling) - **COMPLETED**
**Current Step:** 43.2 (Intent Classification)

## Status
- **Step 43.1 Complete**: Created `internal/chat/types.go` with `Intent`, `ChatRequest`, and `ChatResponse` types. CTO Audit APPROVED.
- **Ready for Step 43.2**: Type contracts are established. Next is implementing the KeywordClassifier.

## Context for Step 43.2
The Goal is to build the `KeywordClassifier` (V1 MVP). This is a simple router that maps user message content to `Intent` values.

Requirements:
1. Create `internal/chat/intents.go`.
2. Implement `ClassifyIntent(message string) Intent` function.
3. Use keyword matching for V1 (e.g., "invoice" → `IntentProcessInvoice`).
4. Write unit tests to verify classification logic.

## Key Files
- `internal/chat/intents.go` (NEW)
- `internal/chat/intents_test.go` (NEW)
- `internal/chat/types.go` (Reference for Intent type)

## Next Actions
1. Define `ClassifyIntent` function signature.
2. Implement keyword-based matching logic.
3. Return `IntentUnknown` for unclassified messages.
4. Write table-driven tests.
