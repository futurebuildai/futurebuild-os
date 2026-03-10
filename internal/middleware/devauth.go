package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/colton/futurebuild/pkg/types"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DevAuthMiddleware provides a development-only auth bypass.
// When DEV_AUTH_BYPASS=true is set, it injects hardcoded claims
// for the demo admin user, bypassing Clerk JWT validation.
type DevAuthMiddleware struct {
	db *pgxpool.Pool
}

func NewDevAuthMiddleware(db *pgxpool.Pool) *DevAuthMiddleware {
	return &DevAuthMiddleware{db: db}
}

// IsEnabled returns true if dev auth bypass is active.
func IsDevAuthEnabled() bool {
	return os.Getenv("DEV_AUTH_BYPASS") == "true"
}

// Handler injects demo claims into the request context.
func (m *DevAuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Look up the demo org and admin user from the database
		var orgID, userID, email, name string
		err := m.db.QueryRow(r.Context(),
			`SELECT o.id, u.id, u.email, u.name
			 FROM organizations o
			 JOIN users u ON u.org_id = o.id AND u.role = 'Admin'
			 WHERE o.slug = 'acme-builders-demo'
			 LIMIT 1`).Scan(&orgID, &userID, &email, &name)

		if err != nil {
			slog.Warn("devauth: failed to load demo user, using fallback", "error", err)
			orgID = "00000000-0000-0000-0000-000000000001"
			userID = "00000000-0000-0000-0000-000000000002"
			email = "admin@acme-builders.demo"
			name = "Demo Admin"
		}

		claims := &types.Claims{
			UserID:      userID,
			OrgID:       orgID,
			Role:        types.UserRoleAdmin,
			SubjectType: types.SubjectTypeUser,
			Email:       email,
			Name:        name,
		}

		ctx := context.WithValue(r.Context(), claimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
