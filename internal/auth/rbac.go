// Package auth provides role-based access control (RBAC) for FutureBuild.
// See STEP_81_ROLE_MAPPING.md: Maps UserRoles to granular permission Scopes.
//
// Roles are static and mapped from Clerk organization roles.
// The permission matrix is defined in code (no DB table needed).
package auth

import (
	"github.com/colton/futurebuild/pkg/types"
)

// Scope represents a granular permission action.
// Format: "resource:action" (e.g., "project:create").
type Scope string

const (
	// ScopeAll grants access to everything. Only Admin role has this.
	ScopeAll Scope = "*"

	// Project scopes
	ScopeProjectRead   Scope = "project:read"
	ScopeProjectCreate Scope = "project:create"
	ScopeProjectDelete Scope = "project:delete"

	// Task scopes
	ScopeTaskRead  Scope = "task:read"
	ScopeTaskWrite Scope = "task:write"

	// Budget/Finance scopes
	ScopeBudgetRead    Scope = "budget:read"
	ScopeBudgetApprove Scope = "budget:approve"
	ScopeFinanceEdit   Scope = "finance:edit"

	// Document scopes
	ScopeDocumentRead  Scope = "document:read"
	ScopeDocumentWrite Scope = "document:write"

	// Chat scopes
	ScopeChatRead  Scope = "chat:read"
	ScopeChatWrite Scope = "chat:write"

	// Settings & member management (Admin only)
	ScopeSettingsWrite Scope = "settings:write"
	ScopeMembersManage Scope = "members:manage"
)

// rolePermissions defines the static mapping from UserRole to allowed Scopes.
// Admin has ScopeAll (wildcard). Builder has read/write for projects, tasks,
// documents, and chat. Viewer has read-only access.
//
// Client and Subcontractor roles inherit Viewer permissions (portal users).
// Unexported to prevent external mutation. Use Can() and AllScopes() for access.
var rolePermissions = map[types.UserRole][]Scope{
	types.UserRoleAdmin: {ScopeAll},

	types.UserRoleBuilder: {
		ScopeProjectRead,
		ScopeProjectCreate,
		ScopeTaskRead,
		ScopeTaskWrite,
		ScopeBudgetRead,
		ScopeFinanceEdit,
		ScopeDocumentRead,
		ScopeDocumentWrite,
		ScopeChatRead,
		ScopeChatWrite,
	},

	types.UserRoleViewer: {
		ScopeProjectRead,
		ScopeTaskRead,
		ScopeBudgetRead,
		ScopeDocumentRead,
		ScopeChatRead,
	},

	// Portal roles: read-only by default
	types.UserRoleClient: {
		ScopeProjectRead,
		ScopeTaskRead,
		ScopeBudgetRead,
	},

	types.UserRoleSubcontractor: {
		ScopeProjectRead,
		ScopeTaskRead,
		ScopeTaskWrite, // Subcontractors can update task progress
	},
}

// Can checks whether a role has the given scope permission.
// Returns true if the role has ScopeAll (wildcard) or the specific scope.
func Can(role types.UserRole, scope Scope) bool {
	if scope == "" {
		return false
	}

	scopes, ok := rolePermissions[role]
	if !ok {
		return false
	}

	for _, s := range scopes {
		if s == ScopeAll || s == scope {
			return true
		}
	}

	return false
}

// AllScopes returns all scopes granted to a role.
// Returns nil for unknown roles.
func AllScopes(role types.UserRole) []Scope {
	scopes, ok := rolePermissions[role]
	if !ok {
		return nil
	}
	result := make([]Scope, len(scopes))
	copy(result, scopes)
	return result
}
