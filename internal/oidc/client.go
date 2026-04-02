package oidc

import (
	"time"

	"github.com/zitadel/oidc/v3/pkg/oidc"
	"github.com/zitadel/oidc/v3/pkg/op"
)

// Client implements the op.Client interface.
// Backed by the oidc_clients PostgreSQL table.
type Client struct {
	id             string
	secretHash     string // bcrypt hash, empty for public clients
	name           string
	redirectURIs   []string
	grantTypes     []oidc.GrantType
	responseTypes  []oidc.ResponseType
	authMethod     oidc.AuthMethod
	scopes         []string
	isActive       bool
	devMode        bool
	loginURL       func(string) string
}

// Compile-time interface check.
var _ op.Client = (*Client)(nil)

func (c *Client) GetID() string              { return c.id }
func (c *Client) RedirectURIs() []string     { return c.redirectURIs }
func (c *Client) PostLogoutRedirectURIs() []string { return nil }
func (c *Client) ApplicationType() op.ApplicationType {
	if c.authMethod == oidc.AuthMethodNone {
		return op.ApplicationTypeNative // Public client (PKCE)
	}
	return op.ApplicationTypeWeb
}
func (c *Client) AuthMethod() oidc.AuthMethod       { return c.authMethod }
func (c *Client) ResponseTypes() []oidc.ResponseType { return c.responseTypes }
func (c *Client) GrantTypes() []oidc.GrantType       { return c.grantTypes }
func (c *Client) LoginURL(id string) string {
	if c.loginURL != nil {
		return c.loginURL(id)
	}
	return "/login?authRequestID=" + id
}
func (c *Client) AccessTokenType() op.AccessTokenType {
	return op.AccessTokenTypeJWT // We issue JWT access tokens with custom claims
}
func (c *Client) IDTokenLifetime() time.Duration { return 1 * time.Hour }
func (c *Client) DevMode() bool                  { return c.devMode }
func (c *Client) RestrictAdditionalIdTokenScopes() func(scopes []string) []string {
	return func(scopes []string) []string { return scopes }
}
func (c *Client) RestrictAdditionalAccessTokenScopes() func(scopes []string) []string {
	return func(scopes []string) []string { return scopes }
}
func (c *Client) IsScopeAllowed(scope string) bool {
	return scope == "org" // Our custom "org" scope for org_id, role, plan_tier
}
func (c *Client) IDTokenUserinfoClaimsAssertion() bool { return false }
func (c *Client) ClockSkew() time.Duration             { return 0 }

// newClientFromDB constructs a Client from database row values.
func newClientFromDB(
	clientID, secretHash, name string,
	redirectURIs, grantTypeStrs, responseTypeStrs []string,
	tokenAuth string,
	scopes []string,
	isActive, devMode bool,
	loginURLFunc func(string) string,
) *Client {
	grantTypes := make([]oidc.GrantType, len(grantTypeStrs))
	for i, gt := range grantTypeStrs {
		grantTypes[i] = oidc.GrantType(gt)
	}
	responseTypes := make([]oidc.ResponseType, len(responseTypeStrs))
	for i, rt := range responseTypeStrs {
		responseTypes[i] = oidc.ResponseType(rt)
	}

	var method oidc.AuthMethod
	switch tokenAuth {
	case "client_secret_post":
		method = oidc.AuthMethodPost
	case "client_secret_basic":
		method = oidc.AuthMethodBasic
	default:
		method = oidc.AuthMethodNone
	}

	return &Client{
		id:            clientID,
		secretHash:    secretHash,
		name:          name,
		redirectURIs:  redirectURIs,
		grantTypes:    grantTypes,
		responseTypes: responseTypes,
		authMethod:    method,
		scopes:        scopes,
		isActive:      isActive,
		devMode:       devMode,
		loginURL:      loginURLFunc,
	}
}
