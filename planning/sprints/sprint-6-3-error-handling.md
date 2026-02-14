# Sprint 6.3: Error Handling

> **Epic:** 6 — Production Hardening (Google Audit)
> **Depends On:** None (can run in parallel)
> **Objective:** Users see friendly errors while technical details go to logging. App degrades gracefully when AI services are unavailable.

---

## Sprint Tasks

### Task 6.3.1: Review and Enhance `fb-error-boundary.ts`

**Status:** ⬜ Not Started

**Current State:**
- [fb-error-boundary.ts](file:///home/colton/Desktop/FutureBuild_HQ/XUI/frontend/src/components/base/fb-error-boundary.ts) (93 lines)
- Basic fallback: shows "⚠️ Data Error" with error message
- Property-driven (`has-error`, `error-message`) — parent sets error state
- No connection to external error tracking (Sentry, LogRocket, etc.)
- No categorization of error types

**Gaps to Address:**

1. **No global error catching** — uncaught Promise rejections and render errors are not captured
2. **No error reporting** — errors stay in the browser console only
3. **No retry mechanism** — user must refresh the entire page
4. **No error categorization** — network errors, auth errors, and data errors all look the same

**Atomic Steps:**

1. **Add global error listeners** in `fb-app-shell.ts`:
   ```ts
   window.addEventListener('unhandledrejection', (e) => {
       console.error('[FBApp] Unhandled rejection:', e.reason);
       this._reportError(e.reason, 'promise_rejection');
   });
   window.addEventListener('error', (e) => {
       console.error('[FBApp] Global error:', e.error);
       this._reportError(e.error, 'runtime_error');
   });
   ```

2. **Enhance `fb-error-boundary.ts`:**
   - Add `errorType` property: `'network' | 'auth' | 'data' | 'ai' | 'unknown'`
   - Render different friendly messages per type:
     - Network: "Connection issue — check your internet and try again"
     - Auth: "Session expired — please sign in again" + login button
     - Data: "Something went wrong loading your data" + retry button
     - AI: "Our AI assistant is temporarily unavailable — you can continue in manual mode"
   - Add "Retry" button that re-dispatches the failed action
   - Add "Report Issue" link (optional, sends error context)

3. **Create `error-reporter.ts`** service [NEW]:
   ```ts
   class ErrorReporter {
       report(error: Error, context: Record<string, unknown>): void {
           // In production: send to Sentry/LogRocket
           // In dev: console.error with structured context
           console.error('[ErrorReporter]', {
               message: error.message,
               stack: error.stack,
               ...context,
               url: window.location.href,
               timestamp: new Date().toISOString(),
               userId: store.user$.value?.id,
           });
       }
   }
   export const errorReporter = new ErrorReporter();
   ```

4. **Integrate Sentry** (optional, config-driven):
   - `npm install @sentry/browser`
   - Initialize in `main.ts` with DSN from env var
   - Wire `errorReporter.report()` to `Sentry.captureException()`

---

### Task 6.3.2: Graceful Degradation for AI Service Outages

**Status:** ⬜ Not Started

**Concept:** If the AI service (Vertex/Gemini/Anthropic) is down, the app should revert to "Manual Mode" without crashing. Users can still:
- Create projects manually (form-based, no AI extraction)
- View/edit schedules (no AI recommendations)
- Approve/reject invoices (no AI confidence scores)
- Use the dashboard (no AI-generated feed cards, show last-known state)

**Atomic Steps:**

1. **Backend: AI health check endpoint:**
   ```go
   // GET /api/v1/health/ai
   func HandleAIHealthCheck(w http.ResponseWriter, r *http.Request) {
       status := aiClient.Ping(ctx)
       json.NewEncoder(w).Encode(map[string]any{
           "status": status, // "healthy", "degraded", "unavailable"
           "model": "gemini-2.0-flash",
           "latency_ms": latency,
       })
   }
   ```

2. **Backend: Wrap all AI calls with fallback:**
   ```go
   func (s *InterrogatorService) ProcessMessage(ctx context.Context, msg Message) (*Response, error) {
       resp, err := s.aiClient.Generate(ctx, prompt)
       if err != nil {
           // Log the AI failure
           s.logger.Warn("AI service unavailable, falling back to manual mode", "error", err)
           return &Response{
               Status:  "manual_mode",
               Reply:   "Our AI assistant is temporarily unavailable. You can continue filling in project details manually.",
               ManualMode: true,
           }, nil
       }
       // ... normal processing
   }
   ```

3. **Frontend: Detect manual mode:**
   ```ts
   // onboarding-store.ts
   export const aiAvailable = signal<boolean>(true);
   ```
   - When API returns `manual_mode: true`, set `aiAvailable.value = false`
   - Show banner: "AI assistant is temporarily unavailable. You can continue manually."
   - Hide AI-specific UI (confidence badges, extraction cards)
   - Show traditional form inputs instead

4. **Frontend: Add `ai-status` indicator** to top bar:
   - Green dot: AI healthy
   - Amber dot: AI degraded (slow responses)
   - Red dot: AI unavailable (manual mode)

5. **Test scenarios:**
   - Kill AI service mock → verify app continues to function
   - Verify onboarding works without AI (manual form entry)
   - Verify dashboard shows last-known feed cards
   - Verify invoice editing works without confidence scores

---

## Codebase References

| File | Path | Lines | Notes |
|------|------|-------|-------|
| fb-error-boundary.ts | `frontend/src/components/base/fb-error-boundary.ts` | 93 | Enhance with types and retry |
| fb-app-shell.ts | `frontend/src/components/layout/fb-app-shell.ts` | 783 | Add global error listeners |
| error-reporter.ts | `frontend/src/services/` | [NEW] | Error reporting service |
| onboarding-store.ts | `frontend/src/store/onboarding-store.ts` | 247 | Add `aiAvailable` signal |
| interrogator_service.md | `backend/shadow/internal/service/interrogator_service.md` | 32 | Add fallback handling |

## Verification Plan

- **Manual:** Throw a render error in a component → verify friendly error boundary shows (not white screen)
- **Manual:** Disconnect network → verify "Connection issue" message (not console-only error)
- **Manual:** Expire auth token → verify "Session expired" message with login button
- **Manual:** Block AI API endpoint → verify "AI unavailable" banner and manual mode works
- **Manual:** In manual mode, complete an onboarding flow using form inputs → verify project created
- **Automated:** Integration test: Call onboarding API with AI service mocked as down → verify `manual_mode: true` in response
