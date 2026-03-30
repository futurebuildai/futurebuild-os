# Phase 19 (Next Epoch) Roadmap: ERP Stabilization & A2A Integration

Based on the local End-to-End browser verification of the FutureBuild OS (Execution) and FB-Brain (Connection) systems, the following architectural blueprint addresses identified friction, bugs, and missing affordances.

## 1. UX/UI State Inconsistencies across ERP Domains

### Issue: Infinite Loading States Without Graceful Fallbacks
- **Observation:** The new domain views (`fb-view-corporate`, `fb-view-employees`, and `fb-view-fleet`) suffer from infinite loaders ("Loading...").
- **Root Cause:** Missing backend API endpoints (`/api/v1/corporate/budgets`, `/api/v1/employees`, `/api/v1/fleet`) return 404s, but there are no timeout or error boundary catch mechanisms to display a failure state gracefully.
- **Action Item (Frontend):** Implement network timeouts and error boundaries on API fetch calls to render "Failure to Load" states with retry affordances, rather than locking the user interface indefinitely.

### Issue: ERP Domains Hidden Behind Admin Shell
- **Observation:** The Corporate Financials, HR, and EAM features requested for typical executive users are exclusively rendered inside the `fb-admin-shell` (`/admin/corporate`, `/admin/employees`, `/admin/fleet`).
- **Resource Friction:** Non-admin executive profiles (like Project Executives or standard Users) cannot access the ERP components cleanly, and these views lack global navigation links outside of the admin dashboard.
- **Action Item (Frontend):** Lift ERP domain views out of the exclusive `/admin` router and into the generic `fb-app-shell` routing. Add explicit navigation elements in the top bar or global side navigation for authorized standard users.

### Issue: Missing Action Features
- **Observation:** Actionable requirements like creating/modifying a corporate ledger entry, assigning prevailing wage rates, and attaching fleet to projects were functionally inaccessible because the initial UI load never transitioned to a "ready" view.
- **Action Item (Frontend):** Implement mock states or stub the backend calls in the local development environment so these components can be visually tested and validated independently of the backend's completion.

## 2. API Backbone Implementation Gaps

### Issue: Missing Phase 18 Endpoints
- **Observation:** None of the core ERP API routes are implemented in the Go backend (`cmd/api/main.go` and `internal/api/handlers`). Additionally, the server returns `{"error":"Tenant not found"}` for baseline API calls without a seeded context.
- **Action Item (Backend):** Implement the routes, repositories, and controllers for:
  - `GET /api/v1/corporate/budgets`, `GET /api/v1/corporate/ar-aging`
  - `GET /api/v1/employees`
  - `GET /api/v1/fleet`
  - Automatically map local unauthenticated requests to a default demo tenant to bypass the `Tenant not found` block.

## 3. OS-to-Brain Handshake (A2A) Edge Cases

### Issue: CORS Blocking the Connection Verification
- **Observation:** The `<fb-settings-brain>` Hub renders perfectly, and the Integration Keys generate natively. However, providing the Brain URL (`http://localhost:8081`) leaves the status forever locked in "Connecting...".
- **Root Cause:** The FutureBuild OS frontend attempts an HTTP handshake with `http://localhost:8081/health` (FB-Brain), but FB-Brain's Go backend does not include the standard local CORS headers (specifically blocking `http://localhost:5173` origins).
- **Action Item (Backend - FB-Brain):** Update the `cors` middleware in FB-Brain to dynamically allow requests from the FutureBuild OS frontend origins.

### Issue: Webhooks Failing Silently on Disconnect
- **Observation:** Triggering a procurement action (e.g., Action Card -> Approve Bid) logged the execution event in FutureBuild OS, but no execution log appeared in FB-Brain. 
- **Action Item (System Architecture):** Ensure webhook dispatchers queue their events if the A2A handshake status isn't firmly validated as "Connected", adding a transient retry mechanism (exponential backoff) rather than failing silently.

## Summary Blueprint Prioritization (Epoch 19)
1. **Hotfix:** Add local CORS headers to the FB-Brain Backend.
2. **Hotfix:** Wrap ERP frontend API calls in error boundaries to eliminate infinite loading states and ensure visual fallback.
3. **Core Backend feature:** Architect and wire up the basic CRUD routes for the Corporate, HR, and Fleet Go handlers in `futurebuild-repo`.
4. **Core Frontend feature:** Refactor routing so ERP module access isn't restricted purely to platform super-admins, exposing it securely to Executive organization roles.
5. **System architecture:** Implement resilient webhook queueing between the FutureBuild OS dispatch and FB-Brain receiver.
