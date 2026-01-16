package models

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestInvoiceJSON(t *testing.T) {
	invoice := Invoice{
		ID:          uuid.New(),
		ProjectID:   uuid.New(),
		VendorName:  "Acme Concrete",
		AmountCents: 125050, // $1250.50 in cents
		LineItems: LineItems{
			{
				Description:    "Foundation Pour",
				Quantity:       1,
				UnitPriceCents: 100000, // $1000.00
				TotalCents:     100000,
			},
			{
				Description:    "Delivery Fee",
				Quantity:       1,
				UnitPriceCents: 25050, // $250.50
				TotalCents:     25050,
			},
		},
		DetectedWBSCode: "6.1",
		Status:          InvoiceStatusPending,
	}

	data, err := json.Marshal(invoice)
	if err != nil {
		t.Fatalf("failed to marshal invoice: %v", err)
	}

	var parsed Invoice
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal invoice: %v", err)
	}

	if parsed.VendorName != invoice.VendorName {
		t.Errorf("expected vendor %s, got %s", invoice.VendorName, parsed.VendorName)
	}

	if len(parsed.LineItems) != 2 {
		t.Errorf("expected 2 line items, got %d", len(parsed.LineItems))
	}

	if parsed.LineItems[0].TotalCents != 100000 {
		t.Errorf("expected first line item total 100000 cents, got %d", parsed.LineItems[0].TotalCents)
	}
}

func TestProjectBudgetJSON(t *testing.T) {
	budget := ProjectBudget{
		ID:                   uuid.New(),
		ProjectID:            uuid.New(),
		WBSPhaseID:           "9.x",
		EstimatedAmountCents: 5000000, // $50,000.00
		CommittedAmountCents: 4500000, // $45,000.00
		ActualAmountCents:    1000000, // $10,000.00
	}

	data, err := json.Marshal(budget)
	if err != nil {
		t.Fatalf("failed to marshal budget: %v", err)
	}

	var parsed ProjectBudget
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal budget: %v", err)
	}

	if parsed.WBSPhaseID != budget.WBSPhaseID {
		t.Errorf("expected wbs_phase_id %s, got %s", budget.WBSPhaseID, parsed.WBSPhaseID)
	}
}
