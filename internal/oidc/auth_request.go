package oidc

import (
	"time"

	"github.com/zitadel/oidc/v3/pkg/oidc"
)

// AuthRequest implements the op.AuthRequest interface.
// It represents an in-flight OIDC authorization request persisted in PostgreSQL.
type AuthRequest struct {
	ID_            string
	CreationDate   time.Time
	ApplicationID  string
	CallbackURI    string
	TransferState  string
	Nonce_         string
	Scopes_        []string
	ResponseType_  oidc.ResponseType
	ResponseMode_  oidc.ResponseMode
	CodeChallenge_ *oidc.CodeChallenge

	// Set after authentication
	UserID   string
	Done_    bool
	AuthTime_ time.Time
}

func (a *AuthRequest) GetID() string                          { return a.ID_ }
func (a *AuthRequest) GetACR() string                         { return "" }
func (a *AuthRequest) GetAMR() []string {
	if a.Done_ {
		return []string{"email"} // magic link email authentication
	}
	return nil
}
func (a *AuthRequest) GetAudience() []string                  { return []string{a.ApplicationID} }
func (a *AuthRequest) GetAuthTime() time.Time                 { return a.AuthTime_ }
func (a *AuthRequest) GetClientID() string                    { return a.ApplicationID }
func (a *AuthRequest) GetCodeChallenge() *oidc.CodeChallenge  { return a.CodeChallenge_ }
func (a *AuthRequest) GetNonce() string                       { return a.Nonce_ }
func (a *AuthRequest) GetRedirectURI() string                 { return a.CallbackURI }
func (a *AuthRequest) GetResponseType() oidc.ResponseType     { return a.ResponseType_ }
func (a *AuthRequest) GetResponseMode() oidc.ResponseMode     { return a.ResponseMode_ }
func (a *AuthRequest) GetScopes() []string                    { return a.Scopes_ }
func (a *AuthRequest) GetState() string                       { return a.TransferState }
func (a *AuthRequest) GetSubject() string                     { return a.UserID }
func (a *AuthRequest) Done() bool                             { return a.Done_ }

// authRequestFromOIDC converts the library's AuthRequest into our internal model.
func authRequestFromOIDC(ar *oidc.AuthRequest, userID string) *AuthRequest {
	var challenge *oidc.CodeChallenge
	if ar.CodeChallenge != "" {
		challenge = &oidc.CodeChallenge{
			Challenge: ar.CodeChallenge,
			Method:    ar.CodeChallengeMethod,
		}
	}
	return &AuthRequest{
		CreationDate:   time.Now(),
		ApplicationID:  ar.ClientID,
		CallbackURI:    ar.RedirectURI,
		TransferState:  ar.State,
		Nonce_:         ar.Nonce,
		Scopes_:        ar.Scopes,
		ResponseType_:  ar.ResponseType,
		ResponseMode_:  ar.ResponseMode,
		CodeChallenge_: challenge,
		UserID:         userID,
	}
}
