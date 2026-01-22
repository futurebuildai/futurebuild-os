package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/colton/futurebuild/test/testhelpers"
	"github.com/google/uuid"
)

func TestProjectLifecycle_HappyPath(t *testing.T) {
	// 1. Setup Stack
	stack := testhelpers.NewIntegrationStack(t)

	// Ensure clean state
	ctx := context.Background()
	if err := stack.TruncateAll(ctx); err != nil {
		t.Fatalf("failed to truncate db: %v", err)
	}

	// 2. Define Test Data
	orgID := uuid.New()
	userID := uuid.New()
	projectID := uuid.New() // Pre-generate ID to match request

	// Create Org first (Foreign Key constraint)
	_, err := stack.DB.Exec(ctx, "INSERT INTO organizations (id, name, slug) VALUES ($1, $2, $3)", orgID, "Test Org", "test-org")
	if err != nil {
		t.Fatalf("failed to insert test org: %v", err)
	}

	permitDate := time.Now().Truncate(24 * time.Hour)
	targetDate := time.Now().AddDate(0, 6, 0).Truncate(24 * time.Hour) // 6 months

	newProject := models.Project{
		ID:               projectID,
		Name:             "Integration Test Project",
		Address:          "123 Test Lane",
		OrgID:            orgID,
		PermitIssuedDate: &permitDate,
		TargetEndDate:    &targetDate,
		GSF:              2500,
		Status:           models.ProjectStatusPreconstruction,
	}

	claims := &types.Claims{
		UserID: userID.String(),
		OrgID:  orgID.String(),
		Role:   types.UserRoleBuilder,
	}

	t.Run("CreateProject", func(t *testing.T) {
		payload, _ := json.Marshal(newProject)

		req := httptest.NewRequest(http.MethodPost, "/projects", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(middleware.WithClaims(req.Context(), claims))

		w := httptest.NewRecorder()
		stack.Router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected status 201, got %d. Body: %s", resp.StatusCode, w.Body.String())
		}

		// Verify DB Side Effect
		var dbName string
		err := stack.DB.QueryRow(ctx, "SELECT name FROM projects WHERE id = $1", projectID).Scan(&dbName)
		if err != nil {
			t.Fatalf("failed to query project from DB: %v", err)
		}
		if dbName != newProject.Name {
			t.Errorf("DB name mismatch: got %s, want %s", dbName, newProject.Name)
		}
	})

	t.Run("GetProject", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/projects/"+projectID.String(), nil)
		req = req.WithContext(middleware.WithClaims(req.Context(), claims))

		w := httptest.NewRecorder()
		stack.Router.ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected status 200 for GET, got %d", resp.StatusCode)
		}

		var fetchedProject models.Project
		if err := json.NewDecoder(w.Body).Decode(&fetchedProject); err != nil {
			t.Fatalf("failed to decode GET response: %v", err)
		}

		if fetchedProject.ID != projectID {
			t.Errorf("fetched ID mismatch: got %s, want %s", fetchedProject.ID, projectID)
		}
		if fetchedProject.Name != newProject.Name {
			t.Errorf("fetched name mismatch: got %s, want %s", fetchedProject.Name, newProject.Name)
		}
	})
}
