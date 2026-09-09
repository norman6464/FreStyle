package ticket

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/pkg/fracindex"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// 既定の雛形（sourceSpaceId 未指定のとき）の色。画面の見本と同じ配色を使う
// （状態の枠＝category の色分けと視覚的に対応させる）。
//
// 進行中が 3 つあるので、進むほど濃くなる並びにして一覧で見分けが付くようにする。
const (
	seedColorTodo     = "#5b6b7a"
	seedColorDev      = "#a0661a"
	seedColorReview   = "#8a5a14"
	seedColorVerify   = "#7a5301"
	seedColorReleased = "#2f6b47"

	seedColorDesignType = "#7c3aed"
	seedColorTaskType   = "#2563eb"
	seedColorBugType    = "#9a3b2e"
)

// EnableTicketsForSpaceUseCase はスペースにチケット機能を有効化する。
//
// 「有効化済み」の正本は「初期状態を持つ現役の状態が 1 つある」（HasActiveInitialTicketStatus）。
// 二重の有効化は repository.ErrTicketsAlreadyEnabled を返す。
//
// SourceSpaceID を指定すると、そのスペースの現役の状態・種別をそのまま複製する
// （設計 Ⅵ。同じ構成を別スペースに揃える手段）。指定が無ければ既定の雛形
// （seedStatuses / seedTypes）を作る。どちらも有効化後は管理画面でいつでも
// 編集できる前提なので、ここでの選択は初期値でしかない。
//
// 複製元スペースへの参照権限の確認はこの usecase の責務ではない（handler / 呼び出し側が
// 別途 CheckSpacePermissionUseCase 等で確かめる）。
type EnableTicketsForSpaceUseCase struct {
	repo      repository.TicketRepository
	txManager repository.TxManager
}

func NewEnableTicketsForSpaceUseCase(
	r repository.TicketRepository, txManager repository.TxManager,
) *EnableTicketsForSpaceUseCase {
	return &EnableTicketsForSpaceUseCase{repo: r, txManager: txManager}
}

type EnableTicketsForSpaceInput struct {
	WorkspaceID string
	SpaceID     string
	// SourceSpaceID が nil なら既定の雛形、非 nil ならそのスペースの現役構成を複製する。
	SourceSpaceID *string
}

// EnableTicketsForSpaceOutput はどれだけ作ったかの要約（画面が「N 個の状態・M 個の
// 種別を作成しました」のように出せるように）。
//
// json タグを明示するのは、この型が handler からそのまま JSON で返るため。
// タグが無いと Go の既定でフィールド名がそのまま（大文字始まり）出てしまい、
// ほかの API（domain の構造体は全部 camelCase のタグ付き）と綴りが食い違う。
type EnableTicketsForSpaceOutput struct {
	StatusCount int `json:"statusCount"`
	TypeCount   int `json:"typeCount"`
}

func (u *EnableTicketsForSpaceUseCase) Execute(
	ctx context.Context, in EnableTicketsForSpaceInput,
) (*EnableTicketsForSpaceOutput, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.SpaceID == "" {
		return nil, errors.New("spaceID is required")
	}

	already, err := u.repo.HasActiveInitialTicketStatus(ctx, in.WorkspaceID, in.SpaceID)
	if err != nil {
		return nil, err
	}
	if already {
		return nil, repository.ErrTicketsAlreadyEnabled
	}

	statuses, types, err := u.buildSeed(ctx, in)
	if err != nil {
		return nil, err
	}

	if err := u.txManager.DoInTx(ctx, func(ctx context.Context) error {
		for _, s := range statuses {
			s := s
			s.WorkspaceID, s.SpaceID = in.WorkspaceID, in.SpaceID
			if err := u.repo.InsertTicketStatus(ctx, &s); err != nil {
				return err
			}
		}
		for _, t := range types {
			t := t
			t.WorkspaceID, t.SpaceID = in.WorkspaceID, in.SpaceID
			if err := u.repo.InsertTicketType(ctx, &t); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return &EnableTicketsForSpaceOutput{StatusCount: len(statuses), TypeCount: len(types)}, nil
}

// buildSeed は作る状態・種別の集合を組み立てる（DB へはまだ書かない）。
// 複製元指定があれば ListTicketStatuses/ListTicketTypes（現役のみ）を読み、
// 無ければ既定の雛形を fracindex で採番する。
func (u *EnableTicketsForSpaceUseCase) buildSeed(
	ctx context.Context, in EnableTicketsForSpaceInput,
) ([]domain.TicketStatus, []domain.TicketType, error) {
	if in.SourceSpaceID != nil {
		statuses, err := u.repo.ListTicketStatuses(ctx, in.WorkspaceID, *in.SourceSpaceID, false)
		if err != nil {
			return nil, nil, err
		}
		types, err := u.repo.ListTicketTypes(ctx, in.WorkspaceID, *in.SourceSpaceID, false)
		if err != nil {
			return nil, nil, err
		}
		// SpaceID は複製先へ書き換える（呼び出し元の Execute で行う）。ここでは
		// 複製元から読んだ値をそのまま返す。
		return statuses, types, nil
	}

	statusPos, err := seedPositions(len(seedStatuses))
	if err != nil {
		return nil, nil, err
	}
	statuses := make([]domain.TicketStatus, len(seedStatuses))
	for i, s := range seedStatuses {
		s.Position = statusPos[i]
		statuses[i] = s
	}

	typePos, err := seedPositions(len(seedTypes))
	if err != nil {
		return nil, nil, err
	}
	types := make([]domain.TicketType, len(seedTypes))
	for i, t := range seedTypes {
		t.Position = typePos[i]
		types[i] = t
	}
	return statuses, types, nil
}

// seedStatuses / seedTypes は有効化の既定の雛形（画面の見本と同じ並び）。
//
// Position はここでは決めない（buildSeed が fracindex で採番して埋める）。
// 有効化のあとは管理画面でいつでも足せる・変えられるので、ここでの選択は初期値でしかない。
var seedStatuses = []domain.TicketStatus{
	{Name: "To Do", Category: domain.TicketStatusCategoryTodo, Color: seedColorTodo, IsInitial: true},
	{Name: "開発", Category: domain.TicketStatusCategoryInProgress, Color: seedColorDev},
	{Name: "レビュー中", Category: domain.TicketStatusCategoryInProgress, Color: seedColorReview},
	{Name: "リリース検証", Category: domain.TicketStatusCategoryInProgress, Color: seedColorVerify},
	{Name: "リリース", Category: domain.TicketStatusCategoryDone, Color: seedColorReleased},
}

var seedTypes = []domain.TicketType{
	{Name: "設計", HierarchyLevel: 1, Color: seedColorDesignType},
	{Name: "開発タスク", HierarchyLevel: 0, Color: seedColorTaskType, IsDefault: true},
	{Name: "バグ", HierarchyLevel: 0, Color: seedColorBugType},
}

// seedPositions は n 個ぶんの position を先頭から順に採る。
// fracindex.Between(prev, "") は「prev の次」を返すので、直前の値を渡して数珠つなぎにする。
func seedPositions(n int) ([]string, error) {
	out := make([]string, n)
	prev := ""
	for i := 0; i < n; i++ {
		p, err := fracindex.Between(prev, "")
		if err != nil {
			return nil, err
		}
		out[i] = p
		prev = p
	}
	return out, nil
}
