package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ClerkClient is a focused HTTP client for the Clerk Backend API.
// Used during invite acceptance to create users and assign org memberships.
type ClerkClient struct {
	secretKey  string
	httpClient *http.Client
}

// NewClerkClient creates a new Clerk Backend API client.
// Returns nil if secretKey is empty (Clerk integration disabled).
func NewClerkClient(secretKey string) *ClerkClient {
	if secretKey == "" {
		return nil
	}
	return &ClerkClient{
		secretKey: secretKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// ClerkUser represents the minimal user data returned by Clerk's create user endpoint.
type ClerkUser struct {
	ID             string `json:"id"`
	PrimaryEmailID string `json:"primary_email_address_id"`
}

// CreateUser creates a new user in Clerk via POST /v1/users.
func (c *ClerkClient) CreateUser(ctx context.Context, email, password, firstName, lastName string) (*ClerkUser, error) {
	body := map[string]interface{}{
		"email_address":  []string{email},
		"password":       password,
		"first_name":     firstName,
		"last_name":      lastName,
		"skip_password_checks": false,
	}

	respBody, err := c.doRequest(ctx, http.MethodPost, "https://api.clerk.com/v1/users", body)
	if err != nil {
		return nil, fmt.Errorf("clerk: create user: %w", err)
	}

	var user ClerkUser
	if err := json.Unmarshal(respBody, &user); err != nil {
		return nil, fmt.Errorf("clerk: decode create user response: %w", err)
	}

	return &user, nil
}

// AddOrgMembership adds a user to a Clerk organization.
// POST /v1/organizations/{orgID}/memberships
func (c *ClerkClient) AddOrgMembership(ctx context.Context, clerkOrgID, userID, role string) error {
	body := map[string]interface{}{
		"user_id": userID,
		"role":    role,
	}

	url := fmt.Sprintf("https://api.clerk.com/v1/organizations/%s/memberships", clerkOrgID)
	_, err := c.doRequest(ctx, http.MethodPost, url, body)
	if err != nil {
		return fmt.Errorf("clerk: add org membership: %w", err)
	}

	return nil
}

// MapInternalRoleToClerk converts an internal role string to Clerk org role.
// Admin maps to org:admin; everything else maps to org:member.
func MapInternalRoleToClerk(role string) string {
	if role == "Admin" {
		return "org:admin"
	}
	return "org:member"
}

// doRequest performs an authenticated HTTP request to the Clerk API.
func (c *ClerkClient) doRequest(ctx context.Context, method, url string, body interface{}) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		reqBody = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.secretKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}
