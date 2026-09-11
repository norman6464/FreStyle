// Package ticket はチケット（仕事 1 件を追いかける記録）の usecase 層。
// backend/internal/usecase/<domain> の 1 つ（kb とは対等な別の境界。usecase/kb を import しない）。
package ticket

import (
	"context"
	"errors"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
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

// ResolveTicketLocationUseCase は URL の /kb/tickets/{ticketId} から、そのチケットが
// どのワークスペースに属するかを決める。
//
// テナントを確定する前の読み取りなので、**呼び出し側は返ったワークスペースで必ず
// 権限判定を通してから応答に使うこと**（kb の ResolvePageLocationUseCase と同じ約束）。
// チケットの ID は全テナントで一意な uuid なので、引くこと自体は越境にならない。
type ResolveTicketLocationUseCase struct {
	tickets    repository.TicketRepository
	workspaces repository.KnowledgeBaseRepository
}

func NewResolveTicketLocationUseCase(
	tickets repository.TicketRepository, workspaces repository.KnowledgeBaseRepository,
) *ResolveTicketLocationUseCase {
	return &ResolveTicketLocationUseCase{tickets: tickets, workspaces: workspaces}
}

// ResolveTicketLocationOutput は解決したワークスペース。画面は slug を受け取って
// 以降の API 呼び出しに使う（URL にワークスペースを出さない既存の規則）。
type ResolveTicketLocationOutput struct {
	Workspace domain.Workspace
}

func (u *ResolveTicketLocationUseCase) Execute(ctx context.Context, ticketID string) (*ResolveTicketLocationOutput, error) {
	if ticketID == "" {
		return nil, repository.ErrTicketNotFound
	}
	workspaceID, err := u.tickets.FindTicketWorkspaceID(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	ws, err := u.workspaces.FindWorkspaceByID(ctx, workspaceID)
	if err != nil {
		return nil, err
	}
	// 停止中のワークスペースは無いものとして扱う。slug の経路は解決の入口
	// （ResolveWorkspaceUseCase）が同じ判定をしているが、この id の経路はそこを通らない。
	// ここで見ないと、停止しても id さえ控えていれば読み続けられる。
	if !ws.IsActive {
		return nil, repository.ErrTicketNotFound
	}
	return &ResolveTicketLocationOutput{Workspace: *ws}, nil
}
