package domain_test

import (
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/stretchr/testify/assert"
)

func Test_WorkspaceRef_未所属はゼロ値でどのワークスペースとも一致しない(t *testing.T) {
	var zero domain.WorkspaceRef
	id, affiliated := zero.WorkspaceID()
	assert.Equal(t, "", id)
	assert.False(t, affiliated)
	assert.False(t, zero.Matches("ws-1"))
	assert.False(t, zero.Matches(""), "未所属は空文字とも一致しない")

	no := domain.NoWorkspace()
	assert.Equal(t, zero, no, "NoWorkspace はゼロ値と同じ")
}

func Test_WorkspaceRef_所属ありは自分のIDだけと一致する(t *testing.T) {
	ref := domain.WorkspaceRefOf("ws-1")

	id, affiliated := ref.WorkspaceID()
	assert.Equal(t, "ws-1", id)
	assert.True(t, affiliated)

	assert.True(t, ref.Matches("ws-1"))
	assert.False(t, ref.Matches("ws-2"), "別ワークスペースとは一致しない")
	assert.False(t, ref.Matches(""), "空文字とは一致しない")
}

func Test_User_WorkspaceRef(t *testing.T) {
	t.Run("workspace_id が nil なら NoWorkspace", func(t *testing.T) {
		u := domain.User{WorkspaceID: nil}
		assert.Equal(t, domain.NoWorkspace(), u.WorkspaceRef())
	})

	t.Run("workspace_id があれば WorkspaceRefOf", func(t *testing.T) {
		wid := "ws-1"
		u := domain.User{WorkspaceID: &wid}
		assert.Equal(t, domain.WorkspaceRefOf("ws-1"), u.WorkspaceRef())
	})
}
