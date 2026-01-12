package types

import "github.com/golang-jwt/jwt/v5"

// SubjectType defines the identity type in JWT claims.
type SubjectType string

const (
	SubjectTypeUser    SubjectType = "user"
	SubjectTypeContact SubjectType = "contact"
)

// Claims represents the standard JWT claims for FutureBuild.
// REQUIRED per Data Spine and API Spec.
type Claims struct {
	UserID string   `json:"user_id"`
	OrgID  string   `json:"org_id"` // REQUIRED per Data Spine for multi-tenancy
	Role        UserRole    `json:"role"`         // REQUIRED per API Spec for role enforcement
	SubjectType SubjectType `json:"subject_type"` // REQUIRED to distinguish between USERS and CONTACTS
	jwt.RegisteredClaims
}

// Principal represents the unified identity object returned in the auth response.
type Principal struct {
	ID          string      `json:"id"`
	OrgID       string      `json:"org_id"`
	Email       string      `json:"email"`
	Name        string      `json:"name"`
	Role        UserRole    `json:"role"`
	SubjectType SubjectType `json:"subject_type"`
	CreatedAt   string      `json:"created_at"`
}

// TokenResponse is a standard OAuth2-style response.
// See Frontend Scope for Signals-based store requirements.
type TokenResponse struct {
	AccessToken string     `json:"access_token"`
	TokenType   string     `json:"token_type"` // Default: "Bearer"
	ExpiresIn   int64      `json:"expires_in"` // Seconds
	Principal   *Principal `json:"principal"`  // Unified identity context
}
