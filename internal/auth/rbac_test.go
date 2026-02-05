package auth

import (
	"testing"

	"github.com/colton/futurebuild/pkg/types"
)

func TestCan_AdminHasAllScopes(t *testing.T) {
	scopes := []Scope{
		ScopeProjectRead, ScopeProjectCreate, ScopeProjectDelete,
		ScopeTaskRead, ScopeTaskWrite,
		ScopeBudgetRead, ScopeBudgetApprove,
		ScopeDocumentRead, ScopeDocumentWrite,
		ScopeChatRead, ScopeChatWrite,
		ScopeSettingsWrite, ScopeMembersManage,
	}

	for _, scope := range scopes {
		if !Can(types.UserRoleAdmin, scope) {
			t.Errorf("Admin should have scope %s but was denied", scope)
		}
	}
}

func TestCan_BuilderPermissions(t *testing.T) {
	allowed := []Scope{
		ScopeProjectRead, ScopeProjectCreate, ScopeProjectComplete,
		ScopeTaskRead, ScopeTaskWrite,
		ScopeBudgetRead, ScopeFinanceEdit,
		ScopeDocumentRead, ScopeDocumentWrite,
		ScopeChatRead, ScopeChatWrite,
		ScopeSettingsWrite,
	}
	denied := []Scope{
		ScopeProjectDelete,
		ScopeBudgetApprove,
		ScopeMembersManage,
	}

	for _, scope := range allowed {
		if !Can(types.UserRoleBuilder, scope) {
			t.Errorf("Builder should have scope %s but was denied", scope)
		}
	}
	for _, scope := range denied {
		if Can(types.UserRoleBuilder, scope) {
			t.Errorf("Builder should NOT have scope %s but was granted", scope)
		}
	}
}

func TestCan_PMPermissions(t *testing.T) {
	allowed := []Scope{
		ScopeProjectRead, ScopeProjectComplete,
		ScopeTaskRead, ScopeTaskWrite,
		ScopeBudgetRead, ScopeFinanceEdit,
		ScopeDocumentRead, ScopeDocumentWrite,
		ScopeChatRead, ScopeChatWrite,
	}
	denied := []Scope{
		ScopeProjectCreate,
		ScopeProjectDelete,
		ScopeBudgetApprove,
		ScopeSettingsWrite,
		ScopeMembersManage,
	}

	for _, scope := range allowed {
		if !Can(types.UserRolePM, scope) {
			t.Errorf("PM should have scope %s but was denied", scope)
		}
	}
	for _, scope := range denied {
		if Can(types.UserRolePM, scope) {
			t.Errorf("PM should NOT have scope %s but was granted", scope)
		}
	}
}

func TestCan_ViewerReadOnly(t *testing.T) {
	allowed := []Scope{
		ScopeProjectRead,
		ScopeTaskRead,
		ScopeBudgetRead,
		ScopeDocumentRead,
		ScopeChatRead,
	}
	denied := []Scope{
		ScopeProjectCreate, ScopeProjectDelete,
		ScopeTaskWrite,
		ScopeBudgetApprove,
		ScopeDocumentWrite,
		ScopeChatWrite,
		ScopeSettingsWrite,
		ScopeMembersManage,
	}

	for _, scope := range allowed {
		if !Can(types.UserRoleViewer, scope) {
			t.Errorf("Viewer should have scope %s but was denied", scope)
		}
	}
	for _, scope := range denied {
		if Can(types.UserRoleViewer, scope) {
			t.Errorf("Viewer should NOT have scope %s but was granted", scope)
		}
	}
}

func TestCan_SubcontractorPermissions(t *testing.T) {
	allowed := []Scope{ScopeProjectRead, ScopeTaskRead, ScopeTaskWrite}
	denied := []Scope{ScopeProjectCreate, ScopeDocumentWrite, ScopeChatWrite, ScopeSettingsWrite}

	for _, scope := range allowed {
		if !Can(types.UserRoleSubcontractor, scope) {
			t.Errorf("Subcontractor should have scope %s but was denied", scope)
		}
	}
	for _, scope := range denied {
		if Can(types.UserRoleSubcontractor, scope) {
			t.Errorf("Subcontractor should NOT have scope %s but was granted", scope)
		}
	}
}

func TestCan_ClientPermissions(t *testing.T) {
	allowed := []Scope{ScopeProjectRead, ScopeTaskRead, ScopeBudgetRead}
	denied := []Scope{ScopeProjectCreate, ScopeTaskWrite, ScopeChatWrite}

	for _, scope := range allowed {
		if !Can(types.UserRoleClient, scope) {
			t.Errorf("Client should have scope %s but was denied", scope)
		}
	}
	for _, scope := range denied {
		if Can(types.UserRoleClient, scope) {
			t.Errorf("Client should NOT have scope %s but was granted", scope)
		}
	}
}

func TestCan_UnknownRoleDenied(t *testing.T) {
	if Can(types.UserRole("Unknown"), ScopeProjectRead) {
		t.Error("Unknown role should be denied all scopes")
	}
}

func TestCan_EmptyScopeDenied(t *testing.T) {
	// Empty scope must always return false, even for Admin (wildcard).
	// This prevents accidental authorization bypass from unset scope parameters.
	if Can(types.UserRoleAdmin, "") {
		t.Error("Empty scope should be denied even for Admin")
	}
	if Can(types.UserRoleBuilder, "") {
		t.Error("Empty scope should be denied for Builder")
	}
	if Can(types.UserRole(""), ScopeProjectRead) {
		t.Error("Empty role should be denied")
	}
}

func TestAllScopes_ReturnsCorrectScopes(t *testing.T) {
	adminScopes := AllScopes(types.UserRoleAdmin)
	if len(adminScopes) != 1 || adminScopes[0] != ScopeAll {
		t.Errorf("Admin should have exactly [*], got %v", adminScopes)
	}

	builderScopes := AllScopes(types.UserRoleBuilder)
	if len(builderScopes) != 12 {
		t.Errorf("Builder should have 12 scopes, got %d", len(builderScopes))
	}

	pmScopes := AllScopes(types.UserRolePM)
	if len(pmScopes) != 10 {
		t.Errorf("PM should have 10 scopes, got %d", len(pmScopes))
	}

	viewerScopes := AllScopes(types.UserRoleViewer)
	if len(viewerScopes) != 5 {
		t.Errorf("Viewer should have 5 scopes, got %d", len(viewerScopes))
	}
}

func TestAllScopes_UnknownRoleReturnsNil(t *testing.T) {
	scopes := AllScopes(types.UserRole("Unknown"))
	if scopes != nil {
		t.Error("Unknown role should return nil scopes")
	}
}

func TestAllScopes_ReturnsCopy(t *testing.T) {
	scopes := AllScopes(types.UserRoleBuilder)
	scopes[0] = "mutated"
	// Original should be unchanged
	if rolePermissions[types.UserRoleBuilder][0] == "mutated" {
		t.Error("AllScopes should return a copy, not a reference to the original slice")
	}
}
