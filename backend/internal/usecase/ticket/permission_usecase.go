// Package ticket はチケット（仕事 1 件を追いかける記録）の usecase 層。
// backend/internal/usecase/<domain> の 1 つ（kb とは対等な別の境界。usecase/kb を import しない）。
package ticket

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// CheckTicketPermissionUseCase は「このユーザーはこのチケットで何ができるか」に答える。
//
// チケットの実効権限はページを介さない「スペース単位」の判定（設計 Ⅳ-H）で、チケット固有の
// 権限テーブルは持たない。既存の KnowledgeBasePermissionRepository.SpacePermissionFactsForUser
// をそのまま使い、grant-role の解決 SQL を二重化しない。ここが担うのは
// 「チケット→そのチケットが属するスペース」の解決（TicketRepository.FindTicket）だけ。
//
// チケットが実在しない・別ワークスペースのものは repository.ErrTicketNotFound をそのまま
// 伝える（handler が 404 にマップする）。権限が無いことと存在しないことを区別しない。
type CheckTicketPermissionUseCase struct {
	tickets repository.TicketRepository
	perms   repository.KnowledgeBasePermissionRepository
}

func NewCheckTicketPermissionUseCase(
	tickets repository.TicketRepository, perms repository.KnowledgeBasePermissionRepository,
) *CheckTicketPermissionUseCase {
	return &CheckTicketPermissionUseCase{tickets: tickets, perms: perms}
}

type CheckTicketPermissionInput struct {
	WorkspaceID string
	TicketID    string
	UserID      uint64
}

func (u *CheckTicketPermissionUseCase) Execute(
	ctx context.Context, in CheckTicketPermissionInput,
) (*domain.ScopePermission, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.TicketID == "" {
		return nil, errors.New("ticketID is required")
	}
	if in.UserID == 0 {
		return nil, errors.New("userID is required")
	}
	t, err := u.tickets.FindTicket(ctx, in.WorkspaceID, in.TicketID)
	if err != nil {
		return nil, err
	}
	facts, err := u.perms.SpacePermissionFactsForUser(ctx, in.WorkspaceID, t.SpaceID, in.UserID)
	if err != nil {
		return nil, err
	}
	perm := domain.ResolveScopePermission(*facts)
	return &perm, nil
}

// ResolveTicketKeyUseCase は表示キー（例 FRESTYLE-12）から ticket_id を解決する。
// 分解できない・非実在はどちらも repository.ErrTicketNotFound に畳む
// （フォーマット違反かどうかで存在の有無が漏れないようにする）。
type ResolveTicketKeyUseCase struct {
	tickets repository.TicketRepository
}

func NewResolveTicketKeyUseCase(tickets repository.TicketRepository) *ResolveTicketKeyUseCase {
	return &ResolveTicketKeyUseCase{tickets: tickets}
}

type ResolveTicketKeyInput struct {
	WorkspaceID string
	Key         string
}

func (u *ResolveTicketKeyUseCase) Execute(ctx context.Context, in ResolveTicketKeyInput) (string, error) {
	if in.WorkspaceID == "" {
		return "", errors.New("workspaceID is required")
	}
	spaceKey, number, ok := domain.ParseTicketKey(in.Key)
	if !ok {
		return "", repository.ErrTicketNotFound
	}
	return u.tickets.ResolveTicketIDByKey(ctx, in.WorkspaceID, spaceKey, number)
}
