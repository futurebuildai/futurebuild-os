package service

import (
	"context"
	"fmt"
	"time"

	"github.com/colton/futurebuild/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// EmployeeService implements EmployeeServicer.
// See BACKEND_SCOPE.md Section 20.2
type EmployeeService struct {
	db *pgxpool.Pool
}

func NewEmployeeService(db *pgxpool.Pool) *EmployeeService {
	return &EmployeeService{db: db}
}

func (s *EmployeeService) CreateEmployee(ctx context.Context, orgID uuid.UUID, emp *models.Employee) error {
	emp.OrgID = orgID
	query := `
		INSERT INTO employees (org_id, contact_id, first_name, last_name, employee_number, email, phone, hire_date, termination_date, status, pay_rate_cents, pay_type, classification)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at`

	return s.db.QueryRow(ctx, query,
		emp.OrgID, emp.ContactID, emp.FirstName, emp.LastName, emp.EmployeeNumber,
		emp.Email, emp.Phone, emp.HireDate, emp.TerminationDate, emp.Status,
		emp.PayRateCents, emp.PayType, emp.Classification,
	).Scan(&emp.ID, &emp.CreatedAt, &emp.UpdatedAt)
}

func (s *EmployeeService) GetEmployee(ctx context.Context, employeeID, orgID uuid.UUID) (*models.Employee, error) {
	query := `
		SELECT id, org_id, contact_id, first_name, last_name, employee_number, email, phone,
			hire_date, termination_date, status, pay_rate_cents, pay_type, classification, created_at, updated_at
		FROM employees
		WHERE id = $1 AND org_id = $2`

	var emp models.Employee
	err := s.db.QueryRow(ctx, query, employeeID, orgID).Scan(
		&emp.ID, &emp.OrgID, &emp.ContactID, &emp.FirstName, &emp.LastName,
		&emp.EmployeeNumber, &emp.Email, &emp.Phone, &emp.HireDate, &emp.TerminationDate,
		&emp.Status, &emp.PayRateCents, &emp.PayType, &emp.Classification,
		&emp.CreatedAt, &emp.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &emp, nil
}

func (s *EmployeeService) ListEmployees(ctx context.Context, orgID uuid.UUID, status string) ([]models.Employee, error) {
	var query string
	var args []interface{}

	if status != "" {
		query = `
			SELECT id, org_id, contact_id, first_name, last_name, employee_number, email, phone,
				hire_date, termination_date, status, pay_rate_cents, pay_type, classification, created_at, updated_at
			FROM employees
			WHERE org_id = $1 AND status = $2
			ORDER BY last_name, first_name`
		args = []interface{}{orgID, status}
	} else {
		query = `
			SELECT id, org_id, contact_id, first_name, last_name, employee_number, email, phone,
				hire_date, termination_date, status, pay_rate_cents, pay_type, classification, created_at, updated_at
			FROM employees
			WHERE org_id = $1
			ORDER BY last_name, first_name`
		args = []interface{}{orgID}
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var employees []models.Employee
	for rows.Next() {
		var emp models.Employee
		if err := rows.Scan(
			&emp.ID, &emp.OrgID, &emp.ContactID, &emp.FirstName, &emp.LastName,
			&emp.EmployeeNumber, &emp.Email, &emp.Phone, &emp.HireDate, &emp.TerminationDate,
			&emp.Status, &emp.PayRateCents, &emp.PayType, &emp.Classification,
			&emp.CreatedAt, &emp.UpdatedAt,
		); err != nil {
			return nil, err
		}
		employees = append(employees, emp)
	}

	return employees, nil
}

func (s *EmployeeService) UpdateEmployee(ctx context.Context, employeeID, orgID uuid.UUID, emp *models.Employee) (*models.Employee, error) {
	query := `
		UPDATE employees SET
			first_name = $3, last_name = $4, employee_number = $5, email = $6,
			phone = $7, hire_date = $8, termination_date = $9, status = $10,
			pay_rate_cents = $11, pay_type = $12, classification = $13
		WHERE id = $1 AND org_id = $2
		RETURNING id, org_id, contact_id, first_name, last_name, employee_number, email, phone,
			hire_date, termination_date, status, pay_rate_cents, pay_type, classification, created_at, updated_at`

	var updated models.Employee
	err := s.db.QueryRow(ctx, query,
		employeeID, orgID, emp.FirstName, emp.LastName, emp.EmployeeNumber,
		emp.Email, emp.Phone, emp.HireDate, emp.TerminationDate, emp.Status,
		emp.PayRateCents, emp.PayType, emp.Classification,
	).Scan(
		&updated.ID, &updated.OrgID, &updated.ContactID, &updated.FirstName, &updated.LastName,
		&updated.EmployeeNumber, &updated.Email, &updated.Phone, &updated.HireDate, &updated.TerminationDate,
		&updated.Status, &updated.PayRateCents, &updated.PayType, &updated.Classification,
		&updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &updated, nil
}

func (s *EmployeeService) LogTime(ctx context.Context, log *models.TimeLog) error {
	query := `
		INSERT INTO time_logs (employee_id, project_id, task_id, log_date, hours_worked, overtime_hours, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`

	return s.db.QueryRow(ctx, query,
		log.EmployeeID, log.ProjectID, log.TaskID, log.LogDate,
		log.HoursWorked, log.OvertimeHours, log.Notes,
	).Scan(&log.ID, &log.CreatedAt, &log.UpdatedAt)
}

func (s *EmployeeService) GetTimeLogs(ctx context.Context, employeeID uuid.UUID) ([]models.TimeLog, error) {
	query := `
		SELECT id, employee_id, project_id, task_id, log_date, hours_worked, overtime_hours,
			notes, approved, approved_by, approved_at, created_at, updated_at
		FROM time_logs
		WHERE employee_id = $1
		ORDER BY log_date DESC`

	rows, err := s.db.Query(ctx, query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []models.TimeLog
	for rows.Next() {
		var l models.TimeLog
		if err := rows.Scan(
			&l.ID, &l.EmployeeID, &l.ProjectID, &l.TaskID, &l.LogDate,
			&l.HoursWorked, &l.OvertimeHours, &l.Notes,
			&l.Approved, &l.ApprovedBy, &l.ApprovedAt,
			&l.CreatedAt, &l.UpdatedAt,
		); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}

	return logs, nil
}

func (s *EmployeeService) ApproveTimeLog(ctx context.Context, logID, approverID uuid.UUID) error {
	query := `
		UPDATE time_logs SET approved = true, approved_by = $2, approved_at = $3
		WHERE id = $1 AND approved = false`

	tag, err := s.db.Exec(ctx, query, logID, approverID, time.Now())
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("time log %s not found or already approved", logID)
	}

	return nil
}

// CalculateLaborBurden computes the deterministic labor cost for a project.
// DETERMINISTIC: SUM(hours * rate) + SUM(overtime * rate * 1.5) from approved time logs.
// Returns total cost in cents. No AI uncertainty — pure arithmetic.
func (s *EmployeeService) CalculateLaborBurden(ctx context.Context, projectID uuid.UUID) (int64, error) {
	query := `
		SELECT COALESCE(SUM(
			CAST(tl.hours_worked * e.pay_rate_cents AS BIGINT) +
			CAST(tl.overtime_hours * e.pay_rate_cents * 1.5 AS BIGINT)
		), 0)
		FROM time_logs tl
		INNER JOIN employees e ON tl.employee_id = e.id
		WHERE tl.project_id = $1 AND tl.approved = true AND e.pay_rate_cents IS NOT NULL`

	var totalCents int64
	err := s.db.QueryRow(ctx, query, projectID).Scan(&totalCents)
	if err != nil {
		return 0, err
	}

	return totalCents, nil
}

func (s *EmployeeService) AddCertification(ctx context.Context, cert *models.Certification) error {
	query := `
		INSERT INTO certifications (employee_id, cert_type, cert_number, issue_date, expiration_date, issuing_authority, document_url, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	return s.db.QueryRow(ctx, query,
		cert.EmployeeID, cert.CertType, cert.CertNumber, cert.IssueDate,
		cert.ExpirationDate, cert.IssuingAuthority, cert.DocumentURL, cert.Status,
	).Scan(&cert.ID, &cert.CreatedAt, &cert.UpdatedAt)
}

func (s *EmployeeService) ListCertifications(ctx context.Context, employeeID uuid.UUID) ([]models.Certification, error) {
	query := `
		SELECT id, employee_id, cert_type, cert_number, issue_date, expiration_date,
			issuing_authority, document_url, status, created_at, updated_at
		FROM certifications
		WHERE employee_id = $1
		ORDER BY expiration_date ASC`

	rows, err := s.db.Query(ctx, query, employeeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var certs []models.Certification
	for rows.Next() {
		var c models.Certification
		if err := rows.Scan(
			&c.ID, &c.EmployeeID, &c.CertType, &c.CertNumber, &c.IssueDate,
			&c.ExpirationDate, &c.IssuingAuthority, &c.DocumentURL, &c.Status,
			&c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		certs = append(certs, c)
	}

	return certs, nil
}

// GetExpiringCertifications finds certifications expiring within the given days for compliance alerts.
func (s *EmployeeService) GetExpiringCertifications(ctx context.Context, orgID uuid.UUID, withinDays int) ([]models.Certification, error) {
	query := `
		SELECT c.id, c.employee_id, c.cert_type, c.cert_number, c.issue_date, c.expiration_date,
			c.issuing_authority, c.document_url, c.status, c.created_at, c.updated_at
		FROM certifications c
		INNER JOIN employees e ON c.employee_id = e.id
		WHERE e.org_id = $1
			AND c.expiration_date <= NOW() + ($2 || ' days')::interval
			AND c.status != 'expired'
		ORDER BY c.expiration_date ASC`

	rows, err := s.db.Query(ctx, query, orgID, fmt.Sprintf("%d", withinDays))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var certs []models.Certification
	for rows.Next() {
		var c models.Certification
		if err := rows.Scan(
			&c.ID, &c.EmployeeID, &c.CertType, &c.CertNumber, &c.IssueDate,
			&c.ExpirationDate, &c.IssuingAuthority, &c.DocumentURL, &c.Status,
			&c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		certs = append(certs, c)
	}

	return certs, nil
}
