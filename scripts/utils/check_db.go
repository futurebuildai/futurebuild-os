package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://localhost:5432"
	}
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, databaseURL)
	if err != nil {
		fmt.Printf("Connection to %s failed: %v\n", databaseURL, err)
		os.Exit(1)
	}
	defer conn.Close(ctx)
	fmt.Printf("Connection to %s succeeded!\n", databaseURL)

	// Verification Queries
	fmt.Println("\n--- VERIFICATION RESULTS ---")

	// 1. Task Count
	var taskCount int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM wbs_tasks").Scan(&taskCount)
	if err != nil {
		fmt.Printf("Failed to count tasks: %v\n", err)
	} else {
		fmt.Printf("Total WBS Tasks: %d\n", taskCount)
	}

	// 2. Phase Count
	var phaseCount int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM wbs_phases").Scan(&phaseCount)
	if err != nil {
		fmt.Printf("Failed to count phases: %v\n", err)
	} else {
		fmt.Printf("Total WBS Phases: %d\n", phaseCount)
	}

	// 3. Inspection Count
	var inspectionCount int
	err = conn.QueryRow(ctx, "SELECT COUNT(*) FROM wbs_tasks WHERE is_inspection = true").Scan(&inspectionCount)
	if err != nil {
		fmt.Printf("Failed to count inspections: %v\n", err)
	} else {
		fmt.Printf("Total Inspection Tasks: %d\n", inspectionCount)
	}

	// 4. Ghost Predecessor Check (WBS 9.3 Roof Framing -> WBS 6.0 Roof Trusses)
	var predecessors []string
	err = conn.QueryRow(ctx, "SELECT predecessor_codes FROM wbs_tasks WHERE code = '9.3'").Scan(&predecessors)
	if err != nil {
		fmt.Printf("Failed to check WBS 9.3: %v\n", err)
	} else {
		found := false
		for _, p := range predecessors {
			if p == "6.0" {
				found = true
				break
			}
		}
		if found {
			fmt.Println("✅ Success: WBS 6.0 (Trusses) is a predecessor to 9.3 (Roof Framing)")
		} else {
			fmt.Printf("❌ Failure: WBS 9.3 predecessors are %v, expected 6.0 to be included\n", predecessors)
		}
	}

	// 6. Ghost Predecessor Check (HVAC 10.1 -> 6.2)
	err = conn.QueryRow(ctx, "SELECT predecessor_codes FROM wbs_tasks WHERE code = '10.1'").Scan(&predecessors)
	if err != nil {
		fmt.Printf("Failed to check WBS 10.1: %v\n", err)
	} else {
		found := false
		for _, p := range predecessors {
			if p == "6.2" {
				found = true
				break
			}
		}
		if found {
			fmt.Println("✅ Success: WBS 6.2 (HVAC) is a predecessor to 10.1")
		} else {
			fmt.Printf("❌ Failure: WBS 10.1 predecessors are %v, expected 6.2\n", predecessors)
		}
	}

	// 7. Ghost Predecessor Check (Electrical 10.2 -> 6.4)
	err = conn.QueryRow(ctx, "SELECT predecessor_codes FROM wbs_tasks WHERE code = '10.2'").Scan(&predecessors)
	if err != nil {
		fmt.Printf("Failed to check WBS 10.2: %v\n", err)
	} else {
		found := false
		for _, p := range predecessors {
			if p == "6.4" {
				found = true
				break
			}
		}
		if found {
			fmt.Println("✅ Success: WBS 6.4 (Electrical) is a predecessor to 10.2")
		} else {
			fmt.Printf("❌ Failure: WBS 10.2 predecessors are %v, expected 6.4\n", predecessors)
		}
	}

	// 8. Ghost Predecessor Check (Garage Doors 13.0 -> 6.7)
	err = conn.QueryRow(ctx, "SELECT predecessor_codes FROM wbs_tasks WHERE code = '13.0'").Scan(&predecessors)
	if err != nil {
		fmt.Printf("Failed to check WBS 13.0: %v\n", err)
	} else {
		found := false
		for _, p := range predecessors {
			if p == "6.7" {
				found = true
				break
			}
		}
		if found {
			fmt.Println("✅ Success: WBS 6.7 (Garage Doors) is a predecessor to 13.0")
		} else {
			fmt.Printf("❌ Failure: WBS 13.0 predecessors are %v, expected 6.7\n", predecessors)
		}
	}

	// 9. Metadata Check (6.0)
	var respParty, deliverable, notes string
	err = conn.QueryRow(ctx, "SELECT responsible_party, deliverable, notes FROM wbs_tasks WHERE code = '6.0'").Scan(&respParty, &deliverable, &notes)
	if err != nil {
		fmt.Printf("Failed to check WBS 6.0 metadata: %v\n", err)
	} else {
		if respParty == "Builder" && deliverable == "Purchase Order" {
			fmt.Printf("✅ Success: WBS 6.0 metadata matches (Resp: %s, Deliv: %s)\n", respParty, deliverable)
		} else {
			fmt.Printf("❌ Failure: WBS 6.0 metadata mismatch (Resp: %s, Deliv: %s)\n", respParty, deliverable)
		}
	}
}
