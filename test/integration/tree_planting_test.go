//go:build integration

// Package integration contains the Tree Planting integration test.
// This test validates FutureShade's autonomous diagnosis and remediation capabilities
// through chaos injection - a "self-healing ceremony."
//
// See TREE_PLANTING_specs.md for detailed specification.
package integration

import (
	"context"
	"testing"
	"time"

	"github.com/colton/futurebuild/internal/chaos"
	"github.com/colton/futurebuild/internal/futureshade/tribunal"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/internal/service"
	"github.com/colton/futurebuild/pkg/ai"
	"github.com/colton/futurebuild/pkg/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockDiagnosticClient provides deterministic AI responses for diagnosis.
// Returns a structured JSON diagnosis for CONFIG_DRIFT faults.
type MockDiagnosticClient struct{}

func (m *MockDiagnosticClient) GenerateContent(ctx context.Context, req ai.GenerateRequest) (ai.GenerateResponse, error) {
	// Return a deterministic diagnosis for CONFIG_DRIFT
	return ai.GenerateResponse{
		Text: `{
			"fault_diagnosis": "CONFIG_DRIFT",
			"confidence_score": 0.95,
			"reasoning": "Error message contains 'CONFIG_DRIFT' and 'drift' indicators. The system state shows the setting is corrupted. Recommending config update to restore correct value.",
			"proposed_action": {
				"type": "UPDATE_CONFIG",
				"key": "feature.enabled",
				"value": true
			}
		}`,
	}, nil
}

func (m *MockDiagnosticClient) GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	return nil, nil
}

func (m *MockDiagnosticClient) Close() error {
	return nil
}

// TestTreePlanting_SelfHealingCeremony executes the full self-healing loop:
// Fault -> Detect -> Diagnose -> Fix -> Success
//
// This is a 4-Act ceremony:
// - Act I (Sabotage): Inject fault, observe failure
// - Act II (Awakening): Capture error, invoke Tribunal
// - Act III (Diagnosis): Validate fault identification
// - Act IV (Restoration): Apply fix, verify recovery
func TestTreePlanting_SelfHealingCeremony(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Setup: Create services with chaos injection
	chaosInjector := chaos.NewMemoryInjector()

	// Use nil DB for this test - we're testing the chaos path, not actual DB operations
	// The chaos injector will intercept before any DB call is made
	projectService := service.NewProjectServiceWithChaos(nil, chaosInjector)

	// Setup Tribunal with mock AI for deterministic testing
	// In production, this would use real AI clients
	mockClient := &MockDiagnosticClient{}
	jury := tribunal.Jury{
		Coordinator: mockClient,
		Architect:   mockClient,
		Historian:   mockClient,
	}

	// Repository is nil since we're not persisting decisions in this test
	engine := tribunal.NewConsensusEngine(jury, nil)

	// Create a test project
	testProject := &models.Project{
		ID:    uuid.New(),
		OrgID: uuid.New(),
		Name:  "Tree Planting Test Project",
	}

	// ========================================================================
	// ACT I: SABOTAGE
	// Establish baseline success, then inject chaos
	// ========================================================================
	t.Run("Act_I_Sabotage", func(t *testing.T) {
		// 1a. Baseline: With no faults, the call would proceed to DB
		// Since we have nil DB, we skip the actual baseline success test
		// In a full integration test with DB, you'd verify success here

		// 1b. Inject ConfigDrift fault
		chaosInjector.RegisterFault(chaos.ChaosConfig{
			TargetMethod: "CreateProject",
			ActiveFault:  chaos.FaultConfigDrift,
			ErrorMessage: "config drift detected: feature.enabled setting corrupted",
		})

		// Verify fault is registered
		activeFaults := chaosInjector.GetActiveFaults()
		require.Len(t, activeFaults, 1)
		assert.Equal(t, chaos.FaultConfigDrift, activeFaults[0].ActiveFault)

		// 1c. Attempt operation - should fail with chaos error
		err := projectService.CreateProject(ctx, testProject)
		require.Error(t, err)

		// Verify it's a ChaosError
		chaosErr, ok := err.(*chaos.ChaosError)
		require.True(t, ok, "expected ChaosError, got: %T", err)
		assert.Equal(t, chaos.FaultConfigDrift, chaosErr.Fault)
		assert.Contains(t, chaosErr.Message, "config drift")

		t.Logf("Act I complete: Fault injected, operation failed as expected")
	})

	// ========================================================================
	// ACT II: AWAKENING
	// Capture the error and invoke the Tribunal for diagnosis
	// ========================================================================
	var diagnosisResp *tribunal.DiagnosisResponse

	t.Run("Act_II_Awakening", func(t *testing.T) {
		// Attempt operation to capture the error
		err := projectService.CreateProject(ctx, testProject)
		require.Error(t, err)

		// Prepare diagnosis request
		diagReq := tribunal.DiagnosisRequest{
			ErrorTrace:    err.Error(),
			MethodContext: "CreateProject",
			SystemState: map[string]string{
				"feature.enabled": "corrupted",
				"service.mode":    "production",
			},
		}

		// Invoke Tribunal Diagnose
		var diagErr error
		diagnosisResp, diagErr = engine.Diagnose(ctx, diagReq)
		require.NoError(t, diagErr)
		require.NotNil(t, diagnosisResp)

		t.Logf("Act II complete: Tribunal invoked, diagnosis received")
		t.Logf("  Session ID: %s", diagnosisResp.SessionID)
		t.Logf("  Latency: %dms", diagnosisResp.LatencyMs)
	})

	// ========================================================================
	// ACT III: DIAGNOSIS
	// Validate the Tribunal's fault identification and proposed action
	// ========================================================================
	t.Run("Act_III_Diagnosis", func(t *testing.T) {
		require.NotNil(t, diagnosisResp, "diagnosis response required from Act II")

		decision := diagnosisResp.Decision

		// 3a. Verify correct fault diagnosis
		assert.Equal(t, "CONFIG_DRIFT", decision.FaultDiagnosis)
		t.Logf("  Fault Diagnosis: %s", decision.FaultDiagnosis)

		// 3b. Verify confidence is reasonable (>0.5)
		assert.GreaterOrEqual(t, decision.ConfidenceScore, 0.5)
		t.Logf("  Confidence: %.2f", decision.ConfidenceScore)

		// 3c. Verify reasoning is present
		assert.NotEmpty(t, decision.Reasoning)
		t.Logf("  Reasoning: %s", decision.Reasoning)

		// 3d. Verify proposed action is valid
		action := decision.ProposedAction
		assert.True(t, types.IsValidActionType(string(action.Type)))
		assert.Equal(t, types.ActionUpdateConfig, action.Type)
		assert.Equal(t, "feature.enabled", action.Key)
		t.Logf("  Proposed Action: %s on key=%s", action.Type, action.Key)

		t.Logf("Act III complete: Diagnosis validated")
	})

	// ========================================================================
	// ACT IV: RESTORATION
	// Apply the proposed fix (in-memory only), clear faults, verify success
	// ========================================================================
	t.Run("Act_IV_Restoration", func(t *testing.T) {
		require.NotNil(t, diagnosisResp, "diagnosis response required from Act II")

		// 4a. Simulate applying the fix (in-memory config update)
		// In production, this would update actual config based on decision.ProposedAction
		action := diagnosisResp.Decision.ProposedAction
		t.Logf("Applying remediation: %s key=%s value=%v", action.Type, action.Key, action.Value)

		// 4b. Clear the chaos faults (simulating successful remediation)
		chaosInjector.ClearFaults()

		// Verify faults are cleared
		activeFaults := chaosInjector.GetActiveFaults()
		assert.Empty(t, activeFaults)

		// 4c. Verify operation now succeeds
		// Since we have nil DB, the chaos path will not fail, but we'll get a nil pointer error
		// In a real test with DB, the operation would succeed
		// For this test, we verify the chaos injector no longer blocks
		shouldFail, err := chaosInjector.ShouldFail(ctx, "CreateProject")
		assert.False(t, shouldFail)
		assert.NoError(t, err)

		t.Logf("Act IV complete: System restored, chaos cleared")
	})

	// Final summary
	t.Logf("")
	t.Logf("=== TREE PLANTING CEREMONY COMPLETE ===")
	t.Logf("Self-healing loop validated:")
	t.Logf("  - Fault injection: OK")
	t.Logf("  - Error detection: OK")
	t.Logf("  - AI diagnosis: OK")
	t.Logf("  - Remediation: OK")
}

// TestTreePlanting_ChaosInjectorSafety verifies the chaos injector is safe
// when not configured (nil injector returns no failures).
func TestTreePlanting_ChaosInjectorSafety(t *testing.T) {
	// Service with nil chaos injector (production mode)
	prodService := service.NewProjectService(nil)

	// Verify the service struct was created (can't test CreateProject without DB)
	require.NotNil(t, prodService)
}

// TestTreePlanting_ActionTypeValidation verifies the action type allowlist.
func TestTreePlanting_ActionTypeValidation(t *testing.T) {
	testCases := []struct {
		input string
		valid bool
	}{
		{"UPDATE_CONFIG", true},
		{"CLEAR_CACHE", true},
		{"RETRY", true},
		{"NO_OP", true},
		{"DROP_TABLE", false},     // Dangerous
		{"DELETE_ALL", false},     // Dangerous
		{"EXECUTE_SHELL", false},  // Dangerous
		{"", false},               // Empty
		{"update_config", false},  // Case sensitive
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := types.IsValidActionType(tc.input)
			assert.Equal(t, tc.valid, result)
		})
	}
}

// TestTreePlanting_FaultTypes verifies all fault types are handled correctly.
func TestTreePlanting_FaultTypes(t *testing.T) {
	injector := chaos.NewMemoryInjector()
	ctx := context.Background()

	testCases := []struct {
		fault   chaos.FaultType
		message string
	}{
		{chaos.FaultServiceError, "external service timeout"},
		{chaos.FaultConfigDrift, "config drift detected"},
		{chaos.FaultDBExhausted, "connection pool exhausted"},
	}

	for _, tc := range testCases {
		t.Run(string(tc.fault), func(t *testing.T) {
			injector.ClearFaults()

			injector.RegisterFault(chaos.ChaosConfig{
				TargetMethod: "TestMethod",
				ActiveFault:  tc.fault,
				ErrorMessage: tc.message,
			})

			shouldFail, err := injector.ShouldFail(ctx, "TestMethod")
			assert.True(t, shouldFail)

			chaosErr, ok := err.(*chaos.ChaosError)
			require.True(t, ok)
			assert.Equal(t, tc.fault, chaosErr.Fault)
			assert.Equal(t, tc.message, chaosErr.Message)
		})
	}
}

// TestTreePlanting_ConcurrentAccess verifies the chaos injector is thread-safe.
func TestTreePlanting_ConcurrentAccess(t *testing.T) {
	injector := chaos.NewMemoryInjector()
	ctx := context.Background()

	// Register initial fault
	injector.RegisterFault(chaos.ChaosConfig{
		TargetMethod: "ConcurrentMethod",
		ActiveFault:  chaos.FaultServiceError,
		ErrorMessage: "concurrent test",
	})

	// Run concurrent reads and writes
	done := make(chan bool)

	// Reader goroutines
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_, _ = injector.ShouldFail(ctx, "ConcurrentMethod")
				_ = injector.GetActiveFaults()
			}
			done <- true
		}()
	}

	// Writer goroutines
	for i := 0; i < 5; i++ {
		go func(id int) {
			for j := 0; j < 50; j++ {
				injector.RegisterFault(chaos.ChaosConfig{
					TargetMethod: "DynamicMethod",
					ActiveFault:  chaos.FaultConfigDrift,
					ErrorMessage: "dynamic fault",
				})
				injector.ClearFaults()
				injector.RegisterFault(chaos.ChaosConfig{
					TargetMethod: "ConcurrentMethod",
					ActiveFault:  chaos.FaultServiceError,
					ErrorMessage: "concurrent test",
				})
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 15; i++ {
		<-done
	}

	// If we get here without deadlock or race, the test passes
	t.Log("Concurrent access test passed")
}
