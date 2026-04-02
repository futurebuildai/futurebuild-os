package oidc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zitadel/oidc/v3/pkg/op"
)

// consentHandler serves the consent UI for the OIDC authorization flow.
// Sprint 0: Auto-approve all consent requests. Will be replaced by Lit UI in Sprint 1.
type consentHandler struct {
	storage  *Storage
	callback func(context.Context, string) string
	issuer   *op.IssuerInterceptor
}

func (c *consentHandler) router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", c.showConsentForm)
	r.Post("/", c.issuer.HandlerFunc(c.handleConsent))
	return r
}

func (c *consentHandler) showConsentForm(w http.ResponseWriter, r *http.Request) {
	authRequestID := r.URL.Query().Get("authRequestID")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>FutureBuild Consent</title>
<style>
  body { font-family: system-ui, sans-serif; max-width: 400px; margin: 80px auto; padding: 0 20px; }
  h1 { color: #1a1a2e; }
  .scopes { background: #f3f4f6; padding: 16px; border-radius: 8px; margin: 16px 0; }
  .scope { padding: 4px 0; }
  .actions { display: flex; gap: 12px; margin-top: 16px; }
  button { padding: 12px 24px; border: none; border-radius: 6px; font-size: 16px; cursor: pointer; }
  .accept { background: #2563eb; color: white; }
  .deny { background: #e5e7eb; color: #1a1a2e; }
</style>
</head>
<body>
  <h1>Authorize FutureBuild OS</h1>
  <p>FutureBuild OS is requesting access to your account.</p>
  <div class="scopes">
    <div class="scope">✓ View your profile</div>
    <div class="scope">✓ Access your email</div>
    <div class="scope">✓ View organization details</div>
  </div>
  <form method="POST">
    <input type="hidden" name="authRequestID" value="%s">
    <div class="actions">
      <button type="submit" name="action" value="accept" class="accept">Allow</button>
      <button type="submit" name="action" value="deny" class="deny">Deny</button>
    </div>
  </form>
</body>
</html>`, authRequestID)
}

func (c *consentHandler) handleConsent(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "cannot parse form", http.StatusBadRequest)
		return
	}

	authRequestID := r.FormValue("authRequestID")
	action := r.FormValue("action")

	if action == "deny" {
		http.Error(w, "consent denied", http.StatusForbidden)
		return
	}

	// For Sprint 0, consent is auto-approved by redirecting to the callback
	http.Redirect(w, r, c.callback(r.Context(), authRequestID), http.StatusFound)
}
