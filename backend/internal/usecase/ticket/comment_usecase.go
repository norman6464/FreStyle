package ticket

import (
	"context"
	"errors"
	"log/slog"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// ErrNotCommentAuthor は投稿者本人でも CanManage でもない相手が、発言の編集・削除を
// 試みたときに返す（handler が 403 にマップする）。
var ErrNotCommentAuthor = errors.New("actor is not the comment author")

// CreateTicketCommentUseCase はチケットへ発言（返信を含む）を 1 件作る。
//
// 本文から @mention（ExtractTicketCommentMentions）を拾い、このワークスペースの一員に
// 解決できた相手だけへ ticket_mentioned 通知を送る。加えて、担当が付いていて発言者本人
// でなければ ticket_commented 通知も送る（「ウォッチ中の相手へ」が段の設計だが、
// ウォッチャーの仕組み自体がまだ無いので、当面は担当をその代わりとして扱う）。
type CreateTicketCommentUseCase struct {
	comments repository.TicketCommentRepository
	tickets  repository.TicketRepository
	perms    repository.KnowledgeBasePermissionRepository
	notifs   repository.NotificationRepository
}

func NewCreateTicketCommentUseCase(
	comments repository.TicketCommentRepository,
	tickets repository.TicketRepository,
	perms repository.KnowledgeBasePermissionRepository,
	notifs repository.NotificationRepository,
) *CreateTicketCommentUseCase {
	return &CreateTicketCommentUseCase{comments: comments, tickets: tickets, perms: perms, notifs: notifs}
}

type CreateTicketCommentInput struct {
	WorkspaceID     string
	TicketID        string
	ParentCommentID *string
	AuthorUserID    uint64
	Body            string
}

func (u *CreateTicketCommentUseCase) Execute(ctx context.Context, in CreateTicketCommentInput) (*domain.TicketComment, error) {
	if in.WorkspaceID == "" || in.TicketID == "" {
		return nil, errors.New("workspaceID and ticketID are required")
	}
	if in.AuthorUserID == 0 {
		return nil, errors.New("authorUserID is required")
	}
	if err := domain.ValidateCommentBody(in.Body); err != nil {
		return nil, err
	}
	t, err := u.tickets.FindTicket(ctx, in.WorkspaceID, in.TicketID)
	if err != nil {
		return nil, err
	}
	if in.ParentCommentID != nil {
		// 親発言は同じチケットに属し、削除済みでないことを確認する（別チケット・
		// 削除済みへの返信を生やせてしまう穴を塞ぐ。AddCommentUseCase の
		// GetCommentThread と同じ役割）。
		if _, err := u.comments.FindTicketComment(ctx, in.WorkspaceID, in.TicketID, *in.ParentCommentID); err != nil {
			return nil, err
		}
	}
	c := &domain.TicketComment{
		WorkspaceID: in.WorkspaceID, TicketID: in.TicketID,
		ParentCommentID: in.ParentCommentID, AuthorUserID: in.AuthorUserID, Body: in.Body,
	}
	if err := u.comments.CreateTicketComment(ctx, c); err != nil {
		return nil, err
	}
	u.notify(ctx, in, t)
	return c, nil
}

// notify は通知の作成に失敗しても発言の作成自体は失敗させない（本体の保存は既に成功して
// いるため。kb の最終編集者名の解決と同じ「失敗しても応答は止めない」扱い）。
func (u *CreateTicketCommentUseCase) notify(ctx context.Context, in CreateTicketCommentInput, t *domain.Ticket) {
	var notifs []domain.Notification

	mentioned := ExtractTicketCommentMentions([]byte(in.Body))
	seen := map[uint64]struct{}{in.AuthorUserID: {}} // 自分への通知は作らない
	// 所属確認は 1 件ずつではなく、名指しされた全員をまとめて 1 回の問い合わせで解決する
	// （メンション数だけ逐次 SELECT が飛ぶと、接続プールを 1 要求が占有し続けてしまう）。
	// メンションが無い発言のほうが多いので、その場合は問い合わせ自体を出さない。
	var members map[uint64]bool
	if len(mentioned) > 0 {
		var err error
		members, err = u.perms.IsWorkspaceMemberBulk(ctx, in.WorkspaceID, mentioned)
		if err != nil {
			slog.WarnContext(ctx, "ticket comment: mention membership check failed", "err", err, "ticketId", in.TicketID)
			members = nil
		}
	}
	for _, userID := range mentioned {
		if _, dup := seen[userID]; dup {
			continue
		}
		seen[userID] = struct{}{}
		if !members[userID] {
			continue
		}
		notifs = append(notifs, domain.Notification{
			UserID: userID, Type: domain.NotificationTypeTicketMentioned,
			Title: "チケットで名指しされました", Body: t.Title,
		})
	}

	if assignment, err := u.tickets.FindTicketAssignment(ctx, in.WorkspaceID, in.TicketID); err == nil && assignment != nil {
		// assignee_principal_id は principals.id（担当は principals への複合 FK。設計 Ⅳ-G）。
		// 通知は users.id 宛にしか送れないので、kind='user' の principal だけ解決する
		// （group / space_all 等の代理担当は、通知の宛先解決自体が段 3 の対象外）。
		if principal, err := u.perms.FindPrincipal(ctx, in.WorkspaceID, assignment.AssigneePrincipalID); err == nil &&
			principal != nil && principal.UserID != nil {
			assigneeUserID := *principal.UserID
			if _, dup := seen[assigneeUserID]; !dup {
				notifs = append(notifs, domain.Notification{
					UserID: assigneeUserID, Type: domain.NotificationTypeTicketCommented,
					Title: "担当チケットにコメントが付きました", Body: t.Title,
				})
			}
		}
	}

	if len(notifs) == 0 {
		return
	}
	if err := u.notifs.CreateMany(ctx, notifs); err != nil {
		slog.WarnContext(ctx, "ticket comment: notification create failed", "err", err, "ticketId", in.TicketID)
	}
}

// UpdateTicketCommentUseCase は発言の本文を書き換える。投稿者本人か、スペースの CanManage を
// 持つ相手だけができる（ActorCanManage は handler が権限判定の結果を渡す — usecase 自体は
// 権限リポジトリを持たない設計に揃える）。
type UpdateTicketCommentUseCase struct {
	repo      repository.TicketCommentRepository
	txManager repository.TxManager
}

func NewUpdateTicketCommentUseCase(repo repository.TicketCommentRepository, txManager repository.TxManager) *UpdateTicketCommentUseCase {
	return &UpdateTicketCommentUseCase{repo: repo, txManager: txManager}
}

type UpdateTicketCommentInput struct {
	WorkspaceID    string
	TicketID       string
	CommentID      string
	ActorUserID    uint64
	ActorCanManage bool
	Body           string
}

func (u *UpdateTicketCommentUseCase) Execute(ctx context.Context, in UpdateTicketCommentInput) (*domain.TicketComment, error) {
	if err := domain.ValidateCommentBody(in.Body); err != nil {
		return nil, err
	}
	existing, err := u.repo.FindTicketComment(ctx, in.WorkspaceID, in.TicketID, in.CommentID)
	if err != nil {
		return nil, err
	}
	if existing.AuthorUserID != in.ActorUserID && !in.ActorCanManage {
		return nil, ErrNotCommentAuthor
	}
	var updated *domain.TicketComment
	// 編集前の本文の退避 → 本文の書き換えは 1 つのトランザクションに入れる。片方だけ
	// 成功すると「退避されていない編集」または「編集前が残ったまま本文が古い」という
	// 中間状態が残るため（CreateCommentThreadUseCase と同じ理由）。
	err = u.txManager.DoInTx(ctx, func(ctx context.Context) error {
		if err := u.repo.InsertTicketCommentEdit(ctx, &domain.TicketCommentEdit{
			WorkspaceID: in.WorkspaceID, CommentID: in.CommentID,
			EditorUserID: in.ActorUserID, PreviousBody: existing.Body,
		}); err != nil {
			return err
		}
		updated, err = u.repo.UpdateTicketCommentBody(ctx, in.WorkspaceID, in.TicketID, in.CommentID, in.Body)
		return err
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// DeleteTicketCommentUseCase は発言を「消えたことにする」（deleted_at。物理削除しない —
// 返信がぶら下がっている場合に親を残す必要があるため）。
type DeleteTicketCommentUseCase struct {
	repo repository.TicketCommentRepository
}

func NewDeleteTicketCommentUseCase(repo repository.TicketCommentRepository) *DeleteTicketCommentUseCase {
	return &DeleteTicketCommentUseCase{repo: repo}
}

type DeleteTicketCommentInput struct {
	WorkspaceID    string
	TicketID       string
	CommentID      string
	ActorUserID    uint64
	ActorCanManage bool
}

func (u *DeleteTicketCommentUseCase) Execute(ctx context.Context, in DeleteTicketCommentInput) error {
	existing, err := u.repo.FindTicketComment(ctx, in.WorkspaceID, in.TicketID, in.CommentID)
	if err != nil {
		return err
	}
	if existing.AuthorUserID != in.ActorUserID && !in.ActorCanManage {
		return ErrNotCommentAuthor
	}
	return u.repo.DeleteTicketComment(ctx, in.WorkspaceID, in.TicketID, in.CommentID)
}

// TicketCommentWithReactions は発言 1 件と、その反応の組。
type TicketCommentWithReactions struct {
	Comment   domain.TicketComment
	Reactions []domain.TicketCommentReaction
}

// ListTicketCommentsUseCase はチケットの発言一覧を、それぞれの反応付きで返す。
// 返信は ParentCommentID を見て呼び出し側（handler / 画面）がツリーへ組み立てる想定
// （フラットな時系列のまま返す。並び替えの都合上ここでは木を組まない）。
type ListTicketCommentsUseCase struct {
	repo repository.TicketCommentRepository
}

func NewListTicketCommentsUseCase(repo repository.TicketCommentRepository) *ListTicketCommentsUseCase {
	return &ListTicketCommentsUseCase{repo: repo}
}

func (u *ListTicketCommentsUseCase) Execute(ctx context.Context, workspaceID, ticketID string) ([]TicketCommentWithReactions, error) {
	comments, err := u.repo.ListTicketComments(ctx, workspaceID, ticketID)
	if err != nil {
		return nil, err
	}
	if len(comments) == 0 {
		return []TicketCommentWithReactions{}, nil
	}
	ids := make([]string, 0, len(comments))
	for _, c := range comments {
		ids = append(ids, c.ID)
	}
	reactions, err := u.repo.ListTicketCommentReactions(ctx, workspaceID, ids)
	if err != nil {
		return nil, err
	}
	byComment := make(map[string][]domain.TicketCommentReaction, len(comments))
	for _, r := range reactions {
		byComment[r.CommentID] = append(byComment[r.CommentID], r)
	}
	out := make([]TicketCommentWithReactions, 0, len(comments))
	for _, c := range comments {
		out = append(out, TicketCommentWithReactions{Comment: c, Reactions: byComment[c.ID]})
	}
	return out, nil
}

// ListTicketCommentEditsUseCase は発言 1 件の編集履歴を返す（オンデマンド。一覧応答には
// 含めない — 編集していない発言が大半なので、常に添えると無駄な行き来が増える）。
type ListTicketCommentEditsUseCase struct {
	comments repository.TicketCommentRepository
}

func NewListTicketCommentEditsUseCase(comments repository.TicketCommentRepository) *ListTicketCommentEditsUseCase {
	return &ListTicketCommentEditsUseCase{comments: comments}
}

func (u *ListTicketCommentEditsUseCase) Execute(ctx context.Context, workspaceID, ticketID, commentID string) ([]domain.TicketCommentEdit, error) {
	// 実在確認（他チケット・他テナントの commentId を渡されても 404 にする）。
	if _, err := u.comments.FindTicketComment(ctx, workspaceID, ticketID, commentID); err != nil {
		return nil, err
	}
	return u.comments.ListTicketCommentEdits(ctx, workspaceID, commentID)
}

// AddTicketCommentReactionUseCase / RemoveTicketCommentReactionUseCase は発言への反応の
// 付け外し。付け外しは冪等（AddTicketCommentReaction のドキュメント参照）。
type AddTicketCommentReactionUseCase struct {
	comments repository.TicketCommentRepository
}

func NewAddTicketCommentReactionUseCase(comments repository.TicketCommentRepository) *AddTicketCommentReactionUseCase {
	return &AddTicketCommentReactionUseCase{comments: comments}
}

func (u *AddTicketCommentReactionUseCase) Execute(ctx context.Context, workspaceID, ticketID, commentID string, userID uint64, emoji string) error {
	if !domain.ValidTicketCommentReactionEmoji(emoji) {
		return domain.ErrInvalidTicketCommentReactionEmoji
	}
	// 実在確認（削除済み・他チケットの commentId への反応を防ぐ）。
	if _, err := u.comments.FindTicketComment(ctx, workspaceID, ticketID, commentID); err != nil {
		return err
	}
	return u.comments.AddTicketCommentReaction(ctx, workspaceID, commentID, userID, emoji)
}

type RemoveTicketCommentReactionUseCase struct {
	comments repository.TicketCommentRepository
}

func NewRemoveTicketCommentReactionUseCase(comments repository.TicketCommentRepository) *RemoveTicketCommentReactionUseCase {
	return &RemoveTicketCommentReactionUseCase{comments: comments}
}

func (u *RemoveTicketCommentReactionUseCase) Execute(ctx context.Context, workspaceID, ticketID, commentID string, userID uint64, emoji string) error {
	if _, err := u.comments.FindTicketComment(ctx, workspaceID, ticketID, commentID); err != nil {
		return err
	}
	return u.comments.RemoveTicketCommentReaction(ctx, workspaceID, commentID, userID, emoji)
}
