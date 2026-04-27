package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

type ListLearningReportsUseCase struct{ repo repository.LearningReportRepository }

func NewListLearningReportsUseCase(r repository.LearningReportRepository) *ListLearningReportsUseCase {
	return &ListLearningReportsUseCase{repo: r}
}

func (u *ListLearningReportsUseCase) Execute(ctx context.Context, userID uint64) ([]domain.LearningReport, error) {
	if userID == 0 {
		return nil, errors.New("userID is required")
	}
	return u.repo.ListByUserID(ctx, userID)
}

type RequestLearningReportUseCase struct {
	repo  repository.LearningReportRepository
	queue repository.SqsEnqueuer
}

func NewRequestLearningReportUseCase(r repository.LearningReportRepository, q repository.SqsEnqueuer) *RequestLearningReportUseCase {
	return &RequestLearningReportUseCase{repo: r, queue: q}
}

type RequestLearningReportInput struct {
	UserID     uint64
	PeriodFrom time.Time
	PeriodTo   time.Time
}

func (u *RequestLearningReportUseCase) Execute(ctx context.Context, in RequestLearningReportInput) (*domain.LearningReport, error) {
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	if !in.PeriodTo.After(in.PeriodFrom) {
		return nil, errors.New("periodTo must be after periodFrom")
	}
	r := &domain.LearningReport{
		UserID: in.UserID, PeriodFrom: in.PeriodFrom, PeriodTo: in.PeriodTo,
		Status: domain.LearningReportStatusPending,
	}
	if err := u.repo.Create(ctx, r); err != nil {
		return nil, err
	}
	if err := u.queue.Enqueue(ctx, r.ID); err != nil {
		return nil, err
	}
	return r, nil
}
