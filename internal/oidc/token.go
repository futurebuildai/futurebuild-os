package oidc

import "time"

// Token represents an access token stored in PostgreSQL.
type Token struct {
	ID             string
	ApplicationID  string
	Subject        string
	RefreshTokenID string
	Audience       []string
	Expiration     time.Time
	Scopes         []string
}

// RefreshToken represents a refresh token stored in PostgreSQL.
type RefreshToken struct {
	ID            string
	Token         string
	AuthTime      time.Time
	AMR           []string
	Audience      []string
	UserID        string
	ApplicationID string
	Expiration    time.Time
	Scopes        []string
	AccessToken   string // Token.ID
}

// RefreshTokenRequest wraps RefreshToken to implement op.RefreshTokenRequest.
type RefreshTokenRequest struct {
	*RefreshToken
}

func (r *RefreshTokenRequest) GetAMR() []string       { return r.AMR }
func (r *RefreshTokenRequest) GetAudience() []string   { return r.Audience }
func (r *RefreshTokenRequest) GetAuthTime() time.Time  { return r.AuthTime }
func (r *RefreshTokenRequest) GetClientID() string     { return r.ApplicationID }
func (r *RefreshTokenRequest) GetScopes() []string     { return r.Scopes }
func (r *RefreshTokenRequest) GetSubject() string      { return r.UserID }
func (r *RefreshTokenRequest) SetCurrentScopes(scopes []string) { r.Scopes = scopes }
