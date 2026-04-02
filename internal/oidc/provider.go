package oidc

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"golang.org/x/text/language"

	"github.com/zitadel/oidc/v3/pkg/op"
)

// SetupOIDCProvider creates the OpenID Provider and returns a chi.Router
// with all OIDC endpoints mounted.
func SetupOIDCProvider(issuer string, storage *Storage, cryptoKey [32]byte, logger *slog.Logger, devMode bool) (chi.Router, error) {
	config := &op.Config{
		CryptoKey:                cryptoKey,
		DefaultLogoutRedirectURI: "/logged-out",
		CodeMethodS256:           true,
		AuthMethodPost:           true,
		AuthMethodPrivateKeyJWT:  false,
		GrantTypeRefreshToken:    true,
		RequestObjectSupported:   false,
		SupportedUILocales:       []language.Tag{language.English},
	}

	opts := []op.Option{
		op.WithLogger(logger.WithGroup("oidc")),
	}

	if devMode {
		opts = append(opts, op.WithAllowInsecure())
	}

	provider, err := op.NewOpenIDProvider(issuer, config, storage, opts...)
	if err != nil {
		return nil, fmt.Errorf("creating OIDC provider: %w", err)
	}

	router := chi.NewRouter()

	// Login UI stub (simple HTML form for Sprint 0)
	login := &loginHandler{
		storage:  storage,
		callback: op.AuthCallbackURL(provider),
		issuer:   op.NewIssuerInterceptor(provider.IssuerFromRequest),
	}
	router.Mount("/login", http.StripPrefix("/login", login.router()))

	// Consent UI stub
	consent := &consentHandler{
		storage:  storage,
		callback: op.AuthCallbackURL(provider),
		issuer:   op.NewIssuerInterceptor(provider.IssuerFromRequest),
	}
	router.Mount("/consent", http.StripPrefix("/consent", consent.router()))

	// Logged-out page
	router.HandleFunc("/logged-out", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<!DOCTYPE html><html><body><h1>Signed Out</h1><p>You have been signed out of FutureBuild.</p></body></html>`))
	})

	// Mount the OIDC provider handler at the root
	// This serves: /.well-known/openid-configuration, /authorize, /oauth/token, /keys, /userinfo, /revoke, etc.
	router.Mount("/", provider)

	return router, nil
}

// CryptoKeyFromHex parses a 32-byte hex string into a [32]byte key.
// If empty, generates a deterministic dev key (NOT for production).
func CryptoKeyFromHex(hexStr string) [32]byte {
	if hexStr != "" {
		decoded, err := hex.DecodeString(hexStr)
		if err == nil && len(decoded) == 32 {
			var key [32]byte
			copy(key[:], decoded)
			return key
		}
	}
	// Dev fallback: deterministic key from a known string
	return sha256.Sum256([]byte("futurebuild-brain-dev-key-not-for-production"))
}
