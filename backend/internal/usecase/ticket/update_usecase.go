package ticket

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// UpdateTicketUseCase はチケットの title / doc / type / priority / 日付を書き換える。
//
// PUT 相当（呼び出し側は現在の望ましい値を毎回すべて渡す）。親の変更は別の usecase
// （設計時点では未実装。段 1 の残タスク）に分ける — 親の変更は親子の階層規則・深さ・
// 周期の検査という重い処理を持ち、1 usecase 1 責務に合わない。
type UpdateTicketUseCase struct {
	repo repository.TicketRepository
}

func NewUpdateTicketUseCase(r repository.TicketRepository) *UpdateTicketUseCase {
	return &UpdateTicketUseCase{repo: r}
}

type UpdateTicketInput struct {
	WorkspaceID string
	TicketID    string
	Title       string
	Doc         string
	TypeID      string
	Priority    domain.TicketPriority
	StartDate   *string
	DueDate     *string
	ActorUserID uint64
}

func (u *UpdateTicketUseCase) Execute(ctx context.Context, in UpdateTicketInput) (*domain.Ticket, error) {
	if in.WorkspaceID == "" {
		return nil, errors.New("workspaceID is required")
	}
	if in.TicketID == "" {
		return nil, errors.New("ticketID is required")
	}
	// ticketID はここで 1 度だけ正規化し、以降のすべての呼び出し（repo への各引数・
	// recordChanges・ReplaceTicketTicketLinks の対象）へ同じ値を渡す。ExtractDocRefs が
	// 本文中の ticketRef を uuid.Parse().String() で正規化して取り出す（doc.go の
	// canonicalRefID）のに対し、この URL 由来の ticketID は無加工のまま比較に使われていた。
	// 同じ UUID でも綴り（大文字/小文字）が違えば別の文字列として扱われるため、
	// 自己リンクの防護（リポジトリ層の文字列比較）が綴り違いの自己参照を弾けなかった
	// （不正な形式ならここでは変えず、下流の既存の not-found 処理に委ねる）。
	if parsed, err := uuid.Parse(in.TicketID); err == nil {
		in.TicketID = parsed.String()
	}
	if in.ActorUserID == 0 {
		return nil, errors.New("actorUserID is required")
	}
	// PUT 相当なので TypeID / Title は毎回必須（部分更新の「未指定」を表す値を持たない）。
	// 空のまま repo.UpdateTicket まで進むと kbParseID("") が弾いて ErrTicketNotFound になり、
	// 「タイトルが空」という入力の問題が「チケットが無い」という誤った意味になってしまう。
	if in.TypeID == "" {
		return nil, errors.New("typeID is required")
	}
	if strings.TrimSpace(in.Title) == "" {
		return nil, domain.ErrInvalidTicketName
	}
	if !domain.ValidTicketDateOrder(in.StartDate, in.DueDate) {
		return nil, domain.ErrTicketDateRangeInverted
	}

	before, err := u.repo.FindTicket(ctx, in.WorkspaceID, in.TicketID)
	if err != nil {
		return nil, err
	}

	typeChanged := in.TypeID != before.TypeID
	if typeChanged {
		if err := u.validateTypeChange(ctx, in.WorkspaceID, before, in.TypeID); err != nil {
			return nil, err
		}
	}

	stripped, err := StripDocRefTitles([]byte(in.Doc))
	if err != nil {
		return nil, fmt.Errorf("invalid doc: %w", err)
	}
	plainText := BuildPlainText(stripped)
	pageIDs, ticketIDs, err := ExtractDocRefs(stripped)
	if err != nil {
		return nil, fmt.Errorf("invalid doc: %w", err)
	}

	updated, err := u.repo.UpdateTicket(ctx, in.WorkspaceID, in.TicketID, repository.TicketUpdateFields{
		TypeID:    in.TypeID,
		Title:     in.Title,
		Doc:       stripped,
		PlainText: plainText,
		Priority:  in.Priority,
		StartDate: in.StartDate,
		DueDate:   in.DueDate,
	})
	if err != nil {
		return nil, err
	}

	if err := u.repo.ReplaceTicketPageLinks(ctx, in.WorkspaceID, in.TicketID, pageIDs); err != nil {
		return nil, err
	}
	if err := u.repo.ReplaceTicketTicketLinks(ctx, in.WorkspaceID, in.TicketID, ticketIDs); err != nil {
		return nil, err
	}

	if err := u.recordChanges(ctx, in, before, stripped); err != nil {
		return nil, err
	}
	return updated, nil
}

// validateTypeChange は設計 Ⅳ-D の階層規則を、種別変更の直前に再検証する。
//
//   - 新しい種別が -1（小作業）で、現役の子を持っているなら拒否
//     （-1 はどんな子も持てない。子の個々のレベルを見るまでもなく決まる）
//   - 親を持っているなら、親の種別との組み合わせが規則を満たすか再検証する
func (u *UpdateTicketUseCase) validateTypeChange(
	ctx context.Context, workspaceID string, before *domain.Ticket, newTypeID string,
) error {
	newType, err := u.repo.FindTicketType(ctx, workspaceID, before.SpaceID, newTypeID)
	if err != nil {
		return err
	}
	if newType.HierarchyLevel == -1 {
		childCount, err := u.repo.CountActiveTicketChildren(ctx, workspaceID, before.ID)
		if err != nil {
			return err
		}
		if childCount > 0 {
			return domain.ErrTicketHierarchyRejected
		}
	}
	if before.ParentID != nil {
		parent, err := u.repo.FindTicket(ctx, workspaceID, *before.ParentID)
		if err != nil {
			return err
		}
		parentType, err := u.repo.FindTicketType(ctx, workspaceID, before.SpaceID, parent.TypeID)
		if err != nil {
			return err
		}
		if err := domain.ValidateTicketParentChild(newType.HierarchyLevel, parentType.HierarchyLevel); err != nil {
			return err
		}
	}
	return nil
}

// recordChanges は変わったフィールドだけを 1 グループにまとめて履歴へ書く。
// 何も変わっていなければ InsertTicketChangeGroup 自体を呼ばない（空グループを残さない）。
//
// doc の比較は stripped（pageRef/ticketRef の title を剥がした後）どうしで行う。
// in.Doc（クライアントから届いた生の値）と before.Doc（保存済みで既に剥がされている）を
// そのまま比べると、本文の中身は同じでもクライアントが持っていた参照先の表示用 title の
// 差だけで「変わった」と誤検知する。
func (u *UpdateTicketUseCase) recordChanges(
	ctx context.Context, in UpdateTicketInput, before *domain.Ticket, strippedDoc []byte,
) error {
	var items []domain.TicketChangeItem
	if in.Title != before.Title {
		old, new := before.Title, in.Title
		items = append(items, domain.TicketChangeItem{Field: domain.TicketChangeFieldTitle, OldValue: &old, NewValue: &new})
	}
	if !DocsEqual(before.Doc, strippedDoc) {
		items = append(items, domain.TicketChangeItem{Field: domain.TicketChangeFieldDoc})
	}
	if in.TypeID != before.TypeID {
		old, new := before.TypeID, in.TypeID
		items = append(items, domain.TicketChangeItem{Field: domain.TicketChangeFieldType, OldValue: &old, NewValue: &new})
	}
	if in.Priority != 0 && in.Priority != before.Priority {
		old, new := fmt.Sprint(int(before.Priority)), fmt.Sprint(int(in.Priority))
		items = append(items, domain.TicketChangeItem{Field: domain.TicketChangeFieldPriority, OldValue: &old, NewValue: &new})
	}
	if len(items) == 0 {
		return nil
	}
	return u.repo.InsertTicketChangeGroup(ctx, &domain.TicketChangeGroup{
		WorkspaceID: in.WorkspaceID, TicketID: in.TicketID, ActorUserID: in.ActorUserID, Items: items,
	})
}
