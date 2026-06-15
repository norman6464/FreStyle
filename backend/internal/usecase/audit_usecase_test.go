package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeAuditRepo は AuditRepository の fake。
type fakeAuditRepo struct {
	recorded []domain.AuditEvent
	listRows []domain.AuditEvent
	recErr   error
	listErr  error
}

func (f *fakeAuditRepo) Record(_ context.Context, e *domain.AuditEvent) error {
	if f.recErr != nil {
		return f.recErr
	}
	f.recorded = append(f.recorded, *e)
	return nil
}

func (f *fakeAuditRepo) ListRecent(context.Context, int) ([]domain.AuditEvent, error) {
	return f.listRows, f.listErr
}

func Test_監査記録_入力をイベントに詰めて保存する(t *testing.T) {
	repo := &fakeAuditRepo{}
	uc := usecase.NewRecordAuditEventUseCase(repo)

	err := uc.Execute(context.Background(), usecase.RecordAuditEventInput{
		ActorID: 9, ActorEmail: "admin@x", ActorRole: domain.RoleSuperAdmin,
		Action: "PATCH /admin/companies/:id/active", TargetID: 3,
	})
	require.NoError(t, err)
	require.Len(t, repo.recorded, 1)
	assert.Equal(t, uint64(9), repo.recorded[0].ActorID)
	assert.Equal(t, "admin@x", repo.recorded[0].ActorEmail)
	assert.Equal(t, "PATCH /admin/companies/:id/active", repo.recorded[0].Action)
	assert.Equal(t, uint64(3), repo.recorded[0].TargetID)
}

func Test_監査記録_保存失敗を伝播(t *testing.T) {
	uc := usecase.NewRecordAuditEventUseCase(&fakeAuditRepo{recErr: errors.New("db")})
	assert.Error(t, uc.Execute(context.Background(), usecase.RecordAuditEventInput{}))
}

func Test_監査ログ一覧_新しい順で返す(t *testing.T) {
	repo := &fakeAuditRepo{listRows: []domain.AuditEvent{{ID: 2}, {ID: 1}}}
	uc := usecase.NewListAuditEventsUseCase(repo)
	rows, err := uc.Execute(context.Background())
	require.NoError(t, err)
	require.Len(t, rows, 2)
	assert.Equal(t, uint64(2), rows[0].ID)
}
