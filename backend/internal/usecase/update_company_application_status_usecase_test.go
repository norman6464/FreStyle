package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase"
)

// statusRepo は UpdateStatus に渡された値を記録する fake。
type statusRepo struct {
	gotID     uint64
	gotStatus string
	err       error
}

func (r *statusRepo) Create(context.Context, *domain.CompanyApplication) error     { return nil }
func (r *statusRepo) ListAll(context.Context) ([]domain.CompanyApplication, error) { return nil, nil }
func (r *statusRepo) UpdateStatus(_ context.Context, id uint64, status string) error {
	r.gotID, r.gotStatus = id, status
	return r.err
}

func TestUpdateCompanyApplicationStatus_NormalizesAndUpdates(t *testing.T) {
	repo := &statusRepo{}
	uc := usecase.NewUpdateCompanyApplicationStatusUseCase(repo)
	if err := uc.Execute(context.Background(), 5, "  Approved "); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.gotID != 5 || repo.gotStatus != domain.CompanyApplicationStatusApproved {
		t.Fatalf("expected (5, approved), got (%d, %q)", repo.gotID, repo.gotStatus)
	}
}

func TestUpdateCompanyApplicationStatus_RejectsBadStatus(t *testing.T) {
	uc := usecase.NewUpdateCompanyApplicationStatusUseCase(&statusRepo{})
	if err := uc.Execute(context.Background(), 5, "banana"); !errors.Is(err, usecase.ErrCompanyApplicationBadStatus) {
		t.Fatalf("expected ErrCompanyApplicationBadStatus, got %v", err)
	}
}

func TestUpdateCompanyApplicationStatus_RequiresID(t *testing.T) {
	uc := usecase.NewUpdateCompanyApplicationStatusUseCase(&statusRepo{})
	if err := uc.Execute(context.Background(), 0, "approved"); err == nil {
		t.Fatal("expected error when id == 0")
	}
}

func TestUpdateCompanyApplicationStatus_PropagatesRepoError(t *testing.T) {
	repo := &statusRepo{err: errors.New("db down")}
	uc := usecase.NewUpdateCompanyApplicationStatusUseCase(repo)
	if err := uc.Execute(context.Background(), 5, "rejected"); err == nil {
		t.Fatal("expected repo error to propagate")
	}
}
