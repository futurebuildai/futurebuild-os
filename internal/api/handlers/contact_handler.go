package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/colton/futurebuild/internal/middleware"
	"github.com/colton/futurebuild/internal/models"
	"github.com/colton/futurebuild/pkg/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ContactHandler handles contact CRUD and project assignment endpoints.
// See FRONTEND_V2_SPEC.md §10.3
type ContactHandler struct {
	db *pgxpool.Pool
}

// NewContactHandler creates a new ContactHandler.
func NewContactHandler(db *pgxpool.Pool) *ContactHandler {
	return &ContactHandler{db: db}
}

// --- Request/Response types ---

// CreateContactRequest is the body for POST /api/v1/contacts.
type CreateContactRequest struct {
	Name              string  `json:"name"`
	Phone             *string `json:"phone,omitempty"`
	Email             *string `json:"email,omitempty"`
	Company           *string `json:"company,omitempty"`
	Role              string  `json:"role"`
	ContactPreference string  `json:"contact_preference,omitempty"`
}

// CreateContactResponse includes a matched flag for dedup awareness.
type CreateContactResponse struct {
	Contact models.Contact `json:"contact"`
	Matched bool           `json:"matched"`
}

// BulkCreateContactsRequest is the body for POST /api/v1/contacts/bulk.
type BulkCreateContactsRequest struct {
	Contacts []CreateContactRequest `json:"contacts"`
}

// BulkCreateContactsResponse returns created and duplicate contacts.
type BulkCreateContactsResponse struct {
	Created    []models.Contact `json:"created"`
	Duplicates []models.Contact `json:"duplicates"`
}

// CreateAssignmentRequest is the body for POST /api/v1/projects/:id/assignments.
type CreateAssignmentRequest struct {
	ContactID  string `json:"contact_id"`
	WBSPhaseID string `json:"wbs_phase_id"`
}

// BulkCreateAssignmentsRequest is the body for POST /api/v1/projects/:id/assignments/bulk.
type BulkCreateAssignmentsRequest struct {
	Assignments []CreateAssignmentRequest `json:"assignments"`
}

// AssignmentRow represents a phase with its assigned contact (or null).
type AssignmentRow struct {
	PhaseCode string          `json:"phase_code"`
	PhaseName string          `json:"phase_name"`
	Contact   *models.Contact `json:"contact"`
}

// --- Handlers ---

// ListContacts handles GET /api/v1/contacts.
// Supports optional ?search= query for name/phone/email matching.
func (h *ContactHandler) ListContacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
		return
	}

	search := strings.TrimSpace(r.URL.Query().Get("search"))

	var rows pgx.Rows
	if search != "" {
		pattern := "%" + search + "%"
		rows, err = h.db.Query(ctx, `
			SELECT id, org_id, name, company, phone, email, role, contact_preference,
				created_at, trades, license_number, portal_enabled, source, updated_at
			FROM contacts
			WHERE org_id = $1 AND (
				name ILIKE $2 OR phone ILIKE $2 OR email ILIKE $2 OR company ILIKE $2
			)
			ORDER BY name
			LIMIT 100
		`, orgID, pattern)
	} else {
		rows, err = h.db.Query(ctx, `
			SELECT id, org_id, name, company, phone, email, role, contact_preference,
				created_at, trades, license_number, portal_enabled, source, updated_at
			FROM contacts
			WHERE org_id = $1
			ORDER BY name
			LIMIT 200
		`, orgID)
	}
	if err != nil {
		slog.Error("contacts: list failed", "error", err, "org_id", orgID)
		http.Error(w, "Failed to list contacts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	contacts := []models.Contact{}
	for rows.Next() {
		var c models.Contact
		if err := rows.Scan(
			&c.ID, &c.OrgID, &c.Name, &c.Company, &c.Phone, &c.Email,
			&c.Role, &c.ContactPreference, &c.CreatedAt,
			&c.Trades, &c.LicenseNumber, &c.PortalEnabled, &c.Source, &c.UpdatedAt,
		); err != nil {
			slog.Error("contacts: scan failed", "error", err)
			continue
		}
		contacts = append(contacts, c)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(contacts)
}

// GetContact handles GET /api/v1/contacts/{id}.
func (h *ContactHandler) GetContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
		return
	}
	contactID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid contact ID", http.StatusBadRequest)
		return
	}

	var c models.Contact
	err = h.db.QueryRow(ctx, `
		SELECT id, org_id, name, company, phone, email, role, contact_preference,
			created_at, trades, license_number, address_city, address_state, address_zip,
			website, notes, portal_enabled, source,
			last_contacted_at, avg_response_time_hours, on_time_rate, updated_at
		FROM contacts
		WHERE id = $1 AND org_id = $2
	`, contactID, orgID).Scan(
		&c.ID, &c.OrgID, &c.Name, &c.Company, &c.Phone, &c.Email,
		&c.Role, &c.ContactPreference, &c.CreatedAt,
		&c.Trades, &c.LicenseNumber, &c.AddressCity, &c.AddressState, &c.AddressZip,
		&c.Website, &c.Notes, &c.PortalEnabled, &c.Source,
		&c.LastContactedAt, &c.AvgResponseTimeHours, &c.OnTimeRate, &c.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Contact not found", http.StatusNotFound)
			return
		}
		slog.Error("contacts: get failed", "error", err, "contact_id", contactID)
		http.Error(w, "Failed to get contact", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(c)
}

// CreateContact handles POST /api/v1/contacts.
// Deduplicates by (org_id, phone) and (org_id, email).
func (h *ContactHandler) CreateContact(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var req CreateContactRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if req.Phone == nil && req.Email == nil {
		http.Error(w, "Phone or email is required", http.StatusBadRequest)
		return
	}

	// Deduplication: check if contact with same phone or email exists in org
	existing, matched := h.findExistingContact(ctx, orgID, req.Phone, req.Email)
	if matched {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(CreateContactResponse{Contact: *existing, Matched: true})
		return
	}

	// Infer contact preference from provided fields
	preference := inferContactPreference(req.ContactPreference, req.Phone, req.Email)
	role := req.Role
	if role == "" {
		role = string(models.UserRoleSubcontractor)
	}

	var c models.Contact
	err = h.db.QueryRow(ctx, `
		INSERT INTO contacts (id, org_id, name, company, phone, email, role, contact_preference, source)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'manual')
		RETURNING id, org_id, name, company, phone, email, role, contact_preference, created_at, source
	`, uuid.New(), orgID, req.Name, req.Company, req.Phone, req.Email, role, preference,
	).Scan(&c.ID, &c.OrgID, &c.Name, &c.Company, &c.Phone, &c.Email,
		&c.Role, &c.ContactPreference, &c.CreatedAt, &c.Source)
	if err != nil {
		slog.Error("contacts: create failed", "error", err, "org_id", orgID)
		http.Error(w, "Failed to create contact", http.StatusInternalServerError)
		return
	}

	slog.Info("contacts: created", "contact_id", c.ID, "org_id", orgID, "name", c.Name)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(CreateContactResponse{Contact: c, Matched: false})
}

// BulkCreateContacts handles POST /api/v1/contacts/bulk.
func (h *ContactHandler) BulkCreateContacts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var req BulkCreateContactsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.Contacts) == 0 {
		http.Error(w, "At least one contact is required", http.StatusBadRequest)
		return
	}
	if len(req.Contacts) > 50 {
		http.Error(w, "Maximum 50 contacts per bulk request", http.StatusBadRequest)
		return
	}

	var created []models.Contact
	var duplicates []models.Contact

	for _, cr := range req.Contacts {
		if cr.Name == "" {
			continue
		}

		existing, matched := h.findExistingContact(ctx, orgID, cr.Phone, cr.Email)
		if matched {
			duplicates = append(duplicates, *existing)
			continue
		}

		preference := inferContactPreference(cr.ContactPreference, cr.Phone, cr.Email)
		role := cr.Role
		if role == "" {
			role = string(models.UserRoleSubcontractor)
		}

		var c models.Contact
		err := h.db.QueryRow(ctx, `
			INSERT INTO contacts (id, org_id, name, company, phone, email, role, contact_preference, source)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'imported')
			RETURNING id, org_id, name, company, phone, email, role, contact_preference, created_at, source
		`, uuid.New(), orgID, cr.Name, cr.Company, cr.Phone, cr.Email, role, preference,
		).Scan(&c.ID, &c.OrgID, &c.Name, &c.Company, &c.Phone, &c.Email,
			&c.Role, &c.ContactPreference, &c.CreatedAt, &c.Source)
		if err != nil {
			slog.Error("contacts: bulk create failed for item", "error", err, "name", cr.Name)
			continue
		}
		created = append(created, c)
	}

	if created == nil {
		created = []models.Contact{}
	}
	if duplicates == nil {
		duplicates = []models.Contact{}
	}

	slog.Info("contacts: bulk created", "created", len(created), "duplicates", len(duplicates), "org_id", orgID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(BulkCreateContactsResponse{Created: created, Duplicates: duplicates})
}

// ListAssignments handles GET /api/v1/projects/:id/assignments.
// Returns all WBS phases with their assigned contacts.
func (h *ContactHandler) ListAssignments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
		return
	}
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	// Verify project belongs to org
	var projectExists bool
	_ = h.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM projects WHERE id = $1 AND org_id = $2)`,
		projectID, orgID).Scan(&projectExists)
	if !projectExists {
		http.Error(w, "Project not found", http.StatusNotFound)
		return
	}

	// Get distinct trade phases for this project from WBS codes
	phaseRows, err := h.db.Query(ctx, `
		SELECT DISTINCT
			split_part(wbs_code, '.', 1) AS phase_code,
			MIN(name) AS phase_name
		FROM project_tasks
		WHERE project_id = $1
		GROUP BY phase_code
		ORDER BY phase_code
	`, projectID)
	if err != nil {
		slog.Error("contacts: list phases failed", "error", err, "project_id", projectID)
		http.Error(w, "Failed to list assignments", http.StatusInternalServerError)
		return
	}
	defer phaseRows.Close()

	type phaseInfo struct {
		Code string
		Name string
	}
	var phases []phaseInfo
	for phaseRows.Next() {
		var p phaseInfo
		if err := phaseRows.Scan(&p.Code, &p.Name); err != nil {
			continue
		}
		phases = append(phases, p)
	}

	// Get existing assignments
	assignmentRows, err := h.db.Query(ctx, `
		SELECT pa.wbs_phase_id,
			c.id, c.org_id, c.name, c.company, c.phone, c.email,
			c.role, c.contact_preference, c.created_at
		FROM project_assignments pa
		JOIN contacts c ON c.id = pa.contact_id
		WHERE pa.project_id = $1 AND c.org_id = $2
	`, projectID, orgID)
	if err != nil {
		slog.Error("contacts: list assignments failed", "error", err, "project_id", projectID)
		http.Error(w, "Failed to list assignments", http.StatusInternalServerError)
		return
	}
	defer assignmentRows.Close()

	assignmentMap := make(map[string]*models.Contact)
	for assignmentRows.Next() {
		var phaseID string
		var c models.Contact
		if err := assignmentRows.Scan(
			&phaseID,
			&c.ID, &c.OrgID, &c.Name, &c.Company, &c.Phone, &c.Email,
			&c.Role, &c.ContactPreference, &c.CreatedAt,
		); err != nil {
			continue
		}
		assignmentMap[phaseID] = &c
	}

	// Build response
	result := make([]AssignmentRow, 0, len(phases))
	for _, p := range phases {
		row := AssignmentRow{
			PhaseCode: p.Code,
			PhaseName: getTradeNameForPhase(p.Code),
			Contact:   assignmentMap[p.Code],
		}
		result = append(result, row)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

// CreateAssignment handles POST /api/v1/projects/:id/assignments.
func (h *ContactHandler) CreateAssignment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
		return
	}
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var req CreateAssignmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if req.ContactID == "" || req.WBSPhaseID == "" {
		http.Error(w, "contact_id and wbs_phase_id are required", http.StatusBadRequest)
		return
	}

	contactID, err := uuid.Parse(req.ContactID)
	if err != nil {
		http.Error(w, "Invalid contact_id", http.StatusBadRequest)
		return
	}

	// Verify contact belongs to org
	var contactExists bool
	_ = h.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM contacts WHERE id = $1 AND org_id = $2)`,
		contactID, orgID).Scan(&contactExists)
	if !contactExists {
		http.Error(w, "Contact not found", http.StatusNotFound)
		return
	}

	// Upsert assignment (one contact per project+phase)
	_, err = h.db.Exec(ctx, `
		INSERT INTO project_assignments (id, project_id, contact_id, wbs_phase_id)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (project_id, wbs_phase_id)
		DO UPDATE SET contact_id = EXCLUDED.contact_id
	`, uuid.New(), projectID, contactID, req.WBSPhaseID)
	if err != nil {
		slog.Error("contacts: create assignment failed", "error", err,
			"project_id", projectID, "contact_id", contactID, "phase", req.WBSPhaseID)
		http.Error(w, "Failed to create assignment", http.StatusInternalServerError)
		return
	}

	slog.Info("contacts: assignment created",
		"project_id", projectID, "contact_id", contactID, "phase", req.WBSPhaseID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

// BulkCreateAssignments handles POST /api/v1/projects/:id/assignments/bulk.
func (h *ContactHandler) BulkCreateAssignments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, err := middleware.GetClaims(ctx)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	orgID, err := uuid.Parse(claims.OrgID)
	if err != nil {
		http.Error(w, "Invalid organization", http.StatusInternalServerError)
		return
	}
	projectID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid project ID", http.StatusBadRequest)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, httputil.MaxBodySize)
	var req BulkCreateAssignmentsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.Assignments) == 0 {
		http.Error(w, "At least one assignment is required", http.StatusBadRequest)
		return
	}

	var created int
	for _, a := range req.Assignments {
		contactID, err := uuid.Parse(a.ContactID)
		if err != nil {
			continue
		}

		// Verify contact belongs to org
		var contactExists bool
		_ = h.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM contacts WHERE id = $1 AND org_id = $2)`,
			contactID, orgID).Scan(&contactExists)
		if !contactExists {
			continue
		}

		_, err = h.db.Exec(ctx, `
			INSERT INTO project_assignments (id, project_id, contact_id, wbs_phase_id)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (project_id, wbs_phase_id)
			DO UPDATE SET contact_id = EXCLUDED.contact_id
		`, uuid.New(), projectID, contactID, a.WBSPhaseID)
		if err != nil {
			slog.Error("contacts: bulk assignment failed", "error", err,
				"project_id", projectID, "contact_id", contactID, "phase", a.WBSPhaseID)
			continue
		}
		created++
	}

	slog.Info("contacts: bulk assignments created", "created", created, "project_id", projectID)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]int{"created": created})
}

// --- Helpers ---

func (h *ContactHandler) findExistingContact(ctx context.Context, orgID uuid.UUID, phone, email *string) (*models.Contact, bool) {
	if phone == nil && email == nil {
		return nil, false
	}

	var c models.Contact
	var err error
	if phone != nil && *phone != "" {
		err = h.db.QueryRow(ctx, `
			SELECT id, org_id, name, company, phone, email, role, contact_preference, created_at, source
			FROM contacts WHERE org_id = $1 AND phone = $2
		`, orgID, *phone).Scan(
			&c.ID, &c.OrgID, &c.Name, &c.Company, &c.Phone, &c.Email,
			&c.Role, &c.ContactPreference, &c.CreatedAt, &c.Source)
		if err == nil {
			return &c, true
		}
	}
	if email != nil && *email != "" {
		err = h.db.QueryRow(ctx, `
			SELECT id, org_id, name, company, phone, email, role, contact_preference, created_at, source
			FROM contacts WHERE org_id = $1 AND email = $2
		`, orgID, *email).Scan(
			&c.ID, &c.OrgID, &c.Name, &c.Company, &c.Phone, &c.Email,
			&c.Role, &c.ContactPreference, &c.CreatedAt, &c.Source)
		if err == nil {
			return &c, true
		}
	}
	return nil, false
}

func inferContactPreference(explicit string, phone, email *string) string {
	if explicit != "" {
		return explicit
	}
	hasPhone := phone != nil && *phone != ""
	hasEmail := email != nil && *email != ""
	if hasPhone && hasEmail {
		return string(models.ContactPreferenceSMS) // Default to SMS when both available
	}
	if hasPhone {
		return string(models.ContactPreferenceSMS)
	}
	return string(models.ContactPreferenceEmail)
}

// getTradeNameForPhase maps WBS phase codes to human-readable trade names.
// See FRONTEND_V2_SPEC.md §10.3
func getTradeNameForPhase(code string) string {
	tradeNames := map[string]string{
		"5":  "Permit & Site Prep",
		"6":  "Foundation",
		"7":  "Framing",
		"8":  "Roofing & Exterior",
		"9":  "Rough-Ins (MEP)",
		"10": "Insulation & Drywall",
		"11": "Finishes",
		"12": "Final Inspections",
		"13": "Punch List & Closeout",
	}
	if name, ok := tradeNames[code]; ok {
		return name
	}
	return "Phase " + code
}
