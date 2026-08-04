package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository/repofakes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_監査記録_入力をイベントに詰めて保存する(t *testing.T) {
	// 保存された内容を検証したいので、fake の中でクロージャに退避する。
	var recorded []domain.AuditEvent
	repo := &repofakes.FakeAuditRepository{
		RecordFunc: func(_ context.Context, e *domain.AuditEvent) error {
			recorded = append(recorded, *e)
			return nil
		},
	}
	uc := usecase.NewRecordAuditEventUseCase(repo)

	err := uc.Execute(context.Background(), usecase.RecordAuditEventInput{
		ActorID: 9, ActorEmail: "admin@x", ActorRole: domain.RoleSuperAdmin,
		Action: "PATCH /admin/companies/:id/active", TargetID: 3,
	})
	require.NoError(t, err)
	require.Len(t, recorded, 1)
	assert.Equal(t, uint64(9), recorded[0].ActorID)
	assert.Equal(t, "admin@x", recorded[0].ActorEmail)
	assert.Equal(t, "PATCH /admin/companies/:id/active", recorded[0].Action)
	assert.Equal(t, uint64(3), recorded[0].TargetID)
}

func Test_監査記録_保存失敗を伝播(t *testing.T) {
	// 「何かエラーが出た」ではなく repository のエラーがそのまま伝わることを見る。
	wantErr := errors.New("db")
	repo := &repofakes.FakeAuditRepository{
		RecordFunc: func(context.Context, *domain.AuditEvent) error { return wantErr },
	}
	uc := usecase.NewRecordAuditEventUseCase(repo)
	err := uc.Execute(context.Background(), usecase.RecordAuditEventInput{})
	require.ErrorIs(t, err, wantErr)
}

func Test_監査ログ一覧_新しい順で返す(t *testing.T) {
	repo := &repofakes.FakeAuditRepository{
		ListRecentFunc: func(context.Context, int) ([]domain.AuditEvent, error) {
			return []domain.AuditEvent{{ID: 2}, {ID: 1}}, nil
		},
	}
	uc := usecase.NewListAuditEventsUseCase(repo)
	rows, err := uc.Execute(context.Background())
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, uint64(2), rows[0].ID)
	assert.Equal(t, int64(1), repo.ListRecentCalls.Load())
}
