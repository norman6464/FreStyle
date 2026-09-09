package ticket

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"

	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// GetTicketUseCase はチケット 1 件を、担当（別表）とあわせて取得する。
//
// 担当を一緒に返すのは、詳細画面が必ず担当を出すため（設計 Ⅶ）。repository 側で
// LEFT JOIN しているので問い合わせは 1 回のまま。
//
// 本文（Doc）は保存時に pageRef / ticketRef の title を剥がしてあるので、そのまま返す
// （title=null のまま）。参照先の現在の題名を読み手ごとに解決して埋め込む処理は
// 段 1 の対象外（画面側は id だけを見て遅延解決する想定。ページの
// rewritePageRefTitles に相当する処理は、実際に画面が必要とする段で追加する）。
type GetTicketUseCase struct {
	repo repository.TicketRepository
}

func NewGetTicketUseCase(r repository.TicketRepository) *GetTicketUseCase {
	return &GetTicketUseCase{repo: r}
}

type GetTicketInput struct {
	WorkspaceID string
	TicketID    string
}

func (u *GetTicketUseCase) Execute(ctx context.Context, in GetTicketInput) (*repository.TicketWithAssignee, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.TicketID == "" {
		return nil, errors.New("ticketID is required")
	}
	return u.repo.FindTicketWithAssignee(ctx, in.WorkspaceID, in.TicketID)
}

// GetTicketAssignmentUseCase はチケットの担当だけを引く。
//
// 変更系（作成・更新・状態変更など）の応答も詳細・一覧と同じ形（担当つき）で返したいが、
// それらの usecase は担当を触らないので、handler が応答を組み立てる直前にこれで補う
// （kb が最終編集者の名前を handler で補うのと同じ分担）。担当が居なければ nil を返す。
type GetTicketAssignmentUseCase struct {
	repo repository.TicketRepository
}

func NewGetTicketAssignmentUseCase(r repository.TicketRepository) *GetTicketAssignmentUseCase {
	return &GetTicketAssignmentUseCase{repo: r}
}

func (u *GetTicketAssignmentUseCase) Execute(ctx context.Context, workspaceID, ticketID string) (*domain.TicketAssignment, error) {
	if workspaceID == "" || ticketID == "" {
		return nil, errors.New("workspaceID and ticketID are required")
	}
	return u.repo.FindTicketAssignment(ctx, workspaceID, ticketID)
}
