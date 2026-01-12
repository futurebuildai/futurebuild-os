package service

import (
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/config"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestGenerateJWT_Claims(t *testing.T) {
	secret := "test-secret"
	cfg := &config.Config{
		JWTSecret: secret,
		JWTExpiry: 24 * time.Hour,
	}
	s := NewAuthService(nil, cfg)

	userID := uuid.New()
	orgID := uuid.New()
	user := &models.User{
		ID:    userID,
		OrgID: orgID,
		Email: "test@example.com",
		Name:  "Test User",
		Role:  "Builder",
	}

	resp, err := s.GenerateJWT(user)
	if err != nil {
		t.Fatalf("GenerateJWT failed: %v", err)
	}

	if resp.AccessToken == "" {
		t.Fatal("AccessToken is empty")
	}

	// Parse and verify token claims
	token, err := jwt.ParseWithClaims(resp.AccessToken, &types.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		t.Fatalf("Failed to parse token: %v", err)
	}

	claims, ok := token.Claims.(*types.Claims)
	if !ok || !token.Valid {
		t.Fatal("Invalid token claims")
	}

	// Assertions per Step 22 Refined Plan
	if claims.UserID != userID.String() {
		t.Errorf("expected UserID %s, got %s", userID.String(), claims.UserID)
	}
	if claims.OrgID != orgID.String() {
		t.Errorf("expected OrgID %s, got %s", orgID.String(), claims.OrgID)
	}
	if claims.Role != types.UserRoleBuilder {
		t.Errorf("expected Role %s, got %s", types.UserRoleBuilder, claims.Role)
	}
	if claims.Subject != userID.String() {
		t.Errorf("expected Subject %s, got %s", userID.String(), claims.Subject)
	}
	if claims.Issuer != "futurebuild" {
		t.Errorf("expected Issuer futurebuild, got %s", claims.Issuer)
	}
}
