package ticket

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// AssignTicketUseCase はチケットの担当者を設定する（1 人。既にいれば置き換える）。
//
// 担当に指名できるのは principals（ワークスペース所属の正本）だけで、DB の複合 FK
// （fk_ticket_assignments_principal、kind='user' 固定の生成列込み）が別ワークスペース・
// 非ユーザー主体を機械的に拒む（設計 Ⅳ-G）。ここでは repository.ErrTicketAssigneeNotFound を
// そのまま伝えるだけで、境界の検査を写経しない。
//
// 履歴の表示名（principal の名前）は解決していない（段 1 の既知のギャップ）。
// old/new とも principal_id を値として残す。
type AssignTicketUseCase struct {
	repo repository.TicketRepository
}

func NewAssignTicketUseCase(r repository.TicketRepository) *AssignTicketUseCase {
	return &AssignTicketUseCase{repo: r}
}

type AssignTicketInput struct {
	WorkspaceID         string
	TicketID            string
	AssigneePrincipalID string
	AssignedByUserID    uint64
}

func (u *AssignTicketUseCase) Execute(ctx context.Context, in AssignTicketInput) (*domain.TicketAssignment, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.TicketID == "" {
		return nil, errors.New("ticketID is required")
	}
	if in.AssigneePrincipalID == "" {
		return nil, errors.New("assigneePrincipalID is required")
	}
	if in.AssignedByUserID == 0 {
		return nil, errors.New("assignedByUserID is required")
	}

	before, err := u.repo.FindTicketAssignment(ctx, in.WorkspaceID, in.TicketID)
	if err != nil {
		return nil, err
	}

	assignment := &domain.TicketAssignment{
		WorkspaceID: in.WorkspaceID, TicketID: in.TicketID,
		AssigneePrincipalID: in.AssigneePrincipalID, AssignedByUserID: in.AssignedByUserID,
	}
	if err := u.repo.UpsertTicketAssignment(ctx, assignment); err != nil {
		return nil, err
	}

	var oldValue *string
	if before != nil {
		oldValue = &before.AssigneePrincipalID
	}
	newValue := in.AssigneePrincipalID
	if err := u.repo.InsertTicketChangeGroup(ctx, &domain.TicketChangeGroup{
		WorkspaceID: in.WorkspaceID, TicketID: in.TicketID, ActorUserID: in.AssignedByUserID,
		Items: []domain.TicketChangeItem{
			{Field: domain.TicketChangeFieldAssignee, OldValue: oldValue, NewValue: &newValue},
		},
	}); err != nil {
		return nil, err
	}
	return assignment, nil
}

// UnassignTicketUseCase はチケットの担当を外す。既に担当が無ければ何もしない（履歴も残さない）。
type UnassignTicketUseCase struct {
	repo repository.TicketRepository
}

func NewUnassignTicketUseCase(r repository.TicketRepository) *UnassignTicketUseCase {
	return &UnassignTicketUseCase{repo: r}
}

type UnassignTicketInput struct {
	WorkspaceID string
	TicketID    string
	ActorUserID uint64
}

func (u *UnassignTicketUseCase) Execute(ctx context.Context, in UnassignTicketInput) error {
	if in.WorkspaceID == "" {
		return errors.New("workspaceID is required")
	}
	if in.TicketID == "" {
		return errors.New("ticketID is required")
	}
	if in.ActorUserID == 0 {
		return errors.New("actorUserID is required")
	}

	before, err := u.repo.FindTicketAssignment(ctx, in.WorkspaceID, in.TicketID)
	if err != nil {
		return err
	}
	if before == nil {
		return nil
	}
	if err := u.repo.DeleteTicketAssignment(ctx, in.WorkspaceID, in.TicketID); err != nil {
		return err
	}
	oldValue := before.AssigneePrincipalID
	return u.repo.InsertTicketChangeGroup(ctx, &domain.TicketChangeGroup{
		WorkspaceID: in.WorkspaceID, TicketID: in.TicketID, ActorUserID: in.ActorUserID,
		Items: []domain.TicketChangeItem{
			{Field: domain.TicketChangeFieldAssignee, OldValue: &oldValue, NewValue: nil},
		},
	})
}
