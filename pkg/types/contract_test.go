package types

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

// TestGenerateContractSamples generates JSON samples for core models to be validated against TS schemas.
func TestGenerateContractSamples(t *testing.T) {
	samplesDir := "../../internal/contract_validation/samples"
	if err := os.MkdirAll(samplesDir, 0755); err != nil {
		t.Fatalf("failed to create samples directory: %v", err)
	}

	samples := map[string]interface{}{
		"Forecast": Forecast{
			Date:                     "2026-01-01",
			HighTempC:                25.5,
			LowTempC:                 15.0,
			PrecipitationMM:          2.5,
			PrecipitationProbability: 0.8,
			Conditions:               "Rainy",
		},
		"Contact": Contact{
			ID:      uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			Name:    "John Doe",
			Company: "BuildIt Inc",
			Phone:   "555-0199",
			Email:   "john@example.com",
			Role:    UserRoleBuilder,
		},
		"InvoiceExtraction": InvoiceExtraction{
			Vendor:           "Steel Co",
			Date:             "2026-01-05",
			InvoiceNumber:    "INV-1001",
			TotalAmountCents: 150050, // $1500.50
			SuggestedWBSCode: "6200-LBR",
			Confidence:       0.98,
			LineItems: []InvoiceExtractionItem{
				{
					Description:    "Steel Beams",
					Quantity:       10,
					UnitPriceCents: 15005, // $150.05
					TotalCents:     150050,
				},
			},
		},
		"GanttData": GanttData{
			ProjectID:        uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			CalculatedAt:     "2026-01-06T12:00:00Z",
			ProjectedEndDate: "2026-06-01",
			CriticalPath:     []string{"1.1", "1.2"},
			Tasks: []GanttTask{
				{
					WBSCode:      "1.1",
					Name:         "Foundation",
					Status:       TaskStatusInProgress,
					EarlyStart:   "2026-01-10",
					EarlyFinish:  "2026-01-20",
					DurationDays: 10,
					IsCritical:   true,
				},
			},
		},
	}

	for name, sample := range samples {
		data, err := json.MarshalIndent(sample, "", "  ")
		if err != nil {
			t.Errorf("failed to marshal %s: %v", name, err)
			continue
		}

		path := filepath.Join(samplesDir, fmt.Sprintf("%s.json", name))
		if err := os.WriteFile(path, data, 0644); err != nil {
			t.Errorf("failed to write %s: %v", name, err)
		}
	}
}
