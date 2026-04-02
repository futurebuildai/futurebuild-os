package oidc

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/zitadel/oidc/v3/pkg/op"
)

// loginHandler serves the login UI for the OIDC authorization flow.
// Sprint 0: Simple HTML form with email/password. Will be replaced by Lit UI + magic link.
type loginHandler struct {
	storage  *Storage
	callback func(context.Context, string) string
	issuer   *op.IssuerInterceptor
}

func (l *loginHandler) router() chi.Router {
	r := chi.NewRouter()
	r.Get("/", l.showLoginForm)
	r.Post("/", l.issuer.HandlerFunc(l.handleLogin))
	return r
}

func (l *loginHandler) showLoginForm(w http.ResponseWriter, r *http.Request) {
	authRequestID := r.URL.Query().Get("authRequestID")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head><title>FutureBuild Login</title>
<style>
  body { font-family: system-ui, sans-serif; max-width: 400px; margin: 80px auto; padding: 0 20px; }
  h1 { color: #1a1a2e; }
  form { display: flex; flex-direction: column; gap: 12px; }
  label { font-weight: 600; font-size: 14px; }
  input { padding: 10px; border: 1px solid #ccc; border-radius: 6px; font-size: 16px; }
  button { padding: 12px; background: #2563eb; color: white; border: none; border-radius: 6px; font-size: 16px; cursor: pointer; }
  button:hover { background: #1d4ed8; }
  .note { font-size: 12px; color: #666; margin-top: 8px; }
</style>
</head>
<body>
  <h1>FutureBuild Login</h1>
  <form method="POST">
    <input type="hidden" name="id" value="%s">
    <label for="username">Email</label>
    <input type="email" name="username" id="username" required autofocus>
    <label for="password">Password</label>
    <input type="password" name="password" id="password" required>
    <button type="submit">Sign In</button>
    <p class="note">Sprint 0: Simple email lookup. Magic link auth coming in Sprint 1.</p>
  </form>
</body>
</html>`, authRequestID)
}

func (l *loginHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "cannot parse form", http.StatusBadRequest)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	id := r.FormValue("id")

	if err := l.storage.CheckUsernamePassword(username, password, id); err != nil {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprintf(w, `<!DOCTYPE html>
<html><body>
  <h1>Login Failed</h1>
  <p>%s</p>
  <a href="/login?authRequestID=%s">Try again</a>
</body></html>`, err.Error(), id)
		return
	}

	http.Redirect(w, r, l.callback(r.Context(), id), http.StatusFound)
}
