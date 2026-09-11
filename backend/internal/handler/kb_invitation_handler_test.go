package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// kbInvitee は招待されるが、まだ所属していないユーザー。kbUserID（f.perms.addMember 済み）
// とは別に用意する — 招待の本質は「まだ principal が無い」ことなので、既存メンバーの ID を
// 使うと検証にならない。
const kbInvitee = uint64(900)

func Test_招待API_一覧は未認証を401にする(t *testing.T) {
	f := newKbFixture(kbCanEdit, 0)
	w := f.do(t, http.MethodGet, "/api/v2/kb/invitations", "")
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func Test_招待API_一覧は自分宛のものだけ返す(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbInvitee)
	require.NoError(t, f.perms.InviteWorkspaceMember(context.Background(), kbWorkspaceID, kbInvitee, kbUserID))
	// 他人宛の招待は混ざらない。
	require.NoError(t, f.perms.InviteWorkspaceMember(context.Background(), kbOtherWorkspaceID, kbUserID, kbInvitee))

	w := f.do(t, http.MethodGet, "/api/v2/kb/invitations", "")

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var got []kbInvitationResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	require.Len(t, got, 1)
	assert.Equal(t, kbWorkspaceSlug, got[0].WorkspaceSlug)
	assert.Equal(t, kbUserID, got[0].InvitedByUserID)
}

func Test_招待API_受諾するとメンバーになる(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbInvitee)
	require.NoError(t, f.perms.InviteWorkspaceMember(context.Background(), kbWorkspaceID, kbInvitee, kbUserID))
	require.Nil(t, f.perms.userPrincipal(kbWorkspaceID, kbInvitee), "前提: まだ非メンバー")

	w := f.do(t, http.MethodPost, "/api/v2/kb/invitations/"+kbWorkspaceSlug+"/accept", "")

	require.Equal(t, http.StatusOK, w.Code, w.Body.String())
	var got kbWorkspaceResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &got))
	assert.Equal(t, kbWorkspaceSlug, got.Slug)
	assert.False(t, got.CanManage, "受諾直後は editor（admin ではない）")
	principal := f.perms.userPrincipal(kbWorkspaceID, kbInvitee)
	require.NotNil(t, principal, "受諾すると principal ができる")
	facts, err := f.perms.PagePermissionFactsForUser(context.Background(), kbWorkspaceID, kbRootPageID, kbInvitee)
	require.NoError(t, err)
	assert.True(t, domain.ResolvePagePermission(*facts).CanEdit, "既定 editor が届く")
}

func Test_招待API_受諾は招待されていなければ404(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbInvitee)
	w := f.do(t, http.MethodPost, "/api/v2/kb/invitations/"+kbWorkspaceSlug+"/accept", "")
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func Test_招待API_辞退すると招待が消え非メンバーのまま(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbInvitee)
	require.NoError(t, f.perms.InviteWorkspaceMember(context.Background(), kbWorkspaceID, kbInvitee, kbUserID))

	w := f.do(t, http.MethodPost, "/api/v2/kb/invitations/"+kbWorkspaceSlug+"/decline", "")

	require.Equal(t, http.StatusNoContent, w.Code)
	assert.Nil(t, f.perms.userPrincipal(kbWorkspaceID, kbInvitee), "辞退してもメンバーにはならない")

	// 辞退済みの招待をもう一度受諾しようとしても通らない。
	accept := f.do(t, http.MethodPost, "/api/v2/kb/invitations/"+kbWorkspaceSlug+"/accept", "")
	assert.Equal(t, http.StatusNotFound, accept.Code)
}

func Test_招待API_辞退は招待されていなければ404(t *testing.T) {
	f := newKbFixture(kbCanEdit, kbInvitee)
	w := f.do(t, http.MethodPost, "/api/v2/kb/invitations/"+kbWorkspaceSlug+"/decline", "")
	assert.Equal(t, http.StatusNotFound, w.Code)
}
