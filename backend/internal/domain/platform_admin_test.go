package domain_test

import (
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/stretchr/testify/require"
)

func Test_運営権限のclaim_欠落と空を区別する(t *testing.T) {
	tests := []struct {
		name    string
		present bool
		groups  []string
		want    domain.PlatformAdminClaim
		grant   bool
		decided bool
	}{
		{"claim が無い", false, nil, domain.PlatformAdminClaimAbsent, false, false},
		{"claim があり admin を含む", true, []string{"users", "admin"}, domain.PlatformAdminClaimGranted, true, true},
		{"claim があり admin を含まない", true, []string{"users"}, domain.PlatformAdminClaimRevoked, false, true},
		{"claim があり空", true, []string{}, domain.PlatformAdminClaimRevoked, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domain.PlatformAdminFromGroups(tt.present, tt.groups)
			require.Equal(t, tt.want, got)
			grant, decided := got.Decided()
			require.Equal(t, tt.decided, decided)
			require.Equal(t, tt.grant, grant)
		})
	}
}

func Test_実効役割_運営権限を失った超管理者は最小権限へ倒れる(t *testing.T) {
	tests := []struct {
		name            string
		stored          domain.RoleName
		isPlatformAdmin bool
		want            domain.RoleName
	}{
		{"運営権限が在る super_admin", domain.RoleSuperAdmin, true, domain.RoleSuperAdmin},
		{"運営権限を失った super_admin", domain.RoleSuperAdmin, false, domain.RoleTrainee},
		{"company_admin は影響を受けない", domain.RoleCompanyAdmin, false, domain.RoleCompanyAdmin},
		{"trainee は影響を受けない", domain.RoleTrainee, false, domain.RoleTrainee},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, domain.ResolveEffectiveRole(tt.stored, tt.isPlatformAdmin))
		})
	}
}
