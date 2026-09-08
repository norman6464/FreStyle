package kb

import (
	"context"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// 提案（page_suggestions）は kb パッケージ直下に置く。comment のような独立パッケージには
// しない — 提案はページ本文そのものの書き込み経路の分岐であり、認可も既存の
// CanComment / CanEdit だけで足りる（page_version_usecase.go の doc と同じ判断）。

// CreateSuggestionUseCase は commenter（閲覧+コメントはできるが編集はできない役割）が
// 保存した本文を、blocks へ直接書き込む代わりに提案として積む。
type CreateSuggestionUseCase struct {
	kbRepo      repository.KnowledgeBaseRepository
	versionRepo repository.PageVersionRepository
	suggestions repository.PageSuggestionRepository
}

func NewCreateSuggestionUseCase(
	kbRepo repository.KnowledgeBaseRepository, versionRepo repository.PageVersionRepository, suggestions repository.PageSuggestionRepository,
) *CreateSuggestionUseCase {
	return &CreateSuggestionUseCase{kbRepo: kbRepo, versionRepo: versionRepo, suggestions: suggestions}
}

type CreateSuggestionInput struct {
	WorkspaceID string
	PageID      string
	// Doc は提案後の本文全体（ProseMirror doc の JSON 文字列）。
	Doc          string
	AuthorUserID uint64
}

func (u *CreateSuggestionUseCase) Execute(ctx context.Context, in CreateSuggestionInput) (*domain.PageSuggestion, error) {
	if in.AuthorUserID == 0 {
		return nil, ErrPageEditorRequired
	}
	page, err := u.kbRepo.FindPage(ctx, in.WorkspaceID, in.PageID)
	if err != nil {
		return nil, err
	}
	if page.ArchivedAt != nil {
		return nil, ErrPageArchived
	}
	// ReplacePageBlocksUseCase.Execute と全く同じ検証パイプラインに通す。ここで不正な doc を
	// 弾いておかないと、あとで採用したときに初めて壊れて発覚してしまう
	// （page_usecase.go の ReplacePageBlocksUseCase.Execute 参照）。
	tree, err := parsePageDoc(StripPageRefTitles(in.Doc))
	if err != nil {
		return nil, err
	}
	// flattenPageDoc は木の中の重複 id を新規 UUID へ採番し直す（regenerateBlockIDs と同じ
	// 効能）。戻り値の rows 自体は使わない（提案は blocks へ書き込まない）が、この後の
	// renderPageDoc が同じ木を見るため、先に呼ぶ必要がある。
	if _, err := flattenPageDoc(tree); err != nil {
		return nil, err
	}
	normalized, err := renderPageDoc(tree)
	if err != nil {
		return nil, err
	}

	// 提案した時点のそのページの最新版を BaseSeq にする。まだ版が 1 つも無ければ nil のまま
	// （差分表示はこの版が無いことを前提に、フロント側が「本文なし」として扱う）。
	var baseSeq *int64
	latest, err := u.versionRepo.GetLatestVersion(ctx, in.WorkspaceID, in.PageID)
	if err != nil {
		return nil, err
	}
	if latest != nil {
		seq := latest.Seq
		baseSeq = &seq
	}

	suggestion := &domain.PageSuggestion{
		WorkspaceID:  in.WorkspaceID,
		PageID:       in.PageID,
		BaseSeq:      baseSeq,
		Doc:          normalized,
		AuthorUserID: in.AuthorUserID,
	}
	if err := u.suggestions.Create(ctx, suggestion); err != nil {
		return nil, err
	}
	return suggestion, nil
}

// ListOpenPageSuggestionsUseCase はページの open な提案一覧（created_at 昇順）を返す。
type ListOpenPageSuggestionsUseCase struct {
	suggestions repository.PageSuggestionRepository
}

func NewListOpenPageSuggestionsUseCase(suggestions repository.PageSuggestionRepository) *ListOpenPageSuggestionsUseCase {
	return &ListOpenPageSuggestionsUseCase{suggestions: suggestions}
}

type ListOpenPageSuggestionsInput struct {
	WorkspaceID string
	PageID      string
}

func (u *ListOpenPageSuggestionsUseCase) Execute(ctx context.Context, in ListOpenPageSuggestionsInput) ([]domain.PageSuggestion, error) {
	return u.suggestions.ListOpen(ctx, in.WorkspaceID, in.PageID)
}

// AcceptPageSuggestionUseCase は提案を採用する — 提案の doc を通常の保存経路
// （ReplacePageBlocksUseCase）へそのまま渡し、本文へ反映しつつ版を 1 つ切る。
//
// 提案の解決（Resolve）と本文の書き換え（replaceBlocks.Execute）は 1 つのトランザクションに
// 入れる。Resolve が先に 'accepted' にした後で replaceBlocks.Execute が失敗したら、
// 提案が'accepted'のまま本文だけ古い、という中間状態を作らないため（両方ロールバックする）。
type AcceptPageSuggestionUseCase struct {
	suggestions   repository.PageSuggestionRepository
	replaceBlocks *ReplacePageBlocksUseCase
	txManager     repository.TxManager
}

func NewAcceptPageSuggestionUseCase(
	suggestions repository.PageSuggestionRepository, replaceBlocks *ReplacePageBlocksUseCase, txManager repository.TxManager,
) *AcceptPageSuggestionUseCase {
	return &AcceptPageSuggestionUseCase{suggestions: suggestions, replaceBlocks: replaceBlocks, txManager: txManager}
}

type AcceptSuggestionInput struct {
	WorkspaceID    string
	PageID         string
	SuggestionID   string
	ResolverUserID uint64
}

func (u *AcceptPageSuggestionUseCase) Execute(ctx context.Context, in AcceptSuggestionInput) (*domain.PageSuggestion, error) {
	var resolved *domain.PageSuggestion
	err := u.txManager.DoInTx(ctx, func(ctx context.Context) error {
		s, err := u.suggestions.Resolve(
			ctx, in.WorkspaceID, in.PageID, in.SuggestionID, domain.PageSuggestionStatusAccepted, in.ResolverUserID, time.Now(),
		)
		if err != nil {
			return err
		}
		// ForceVersion は必須 — 「採用すると版が 1 つ切られる」を、10 分規則の間引きを
		// 無視して必ず版を切ることで満たす（「版を残す」・復元と同じ扱い）。
		if _, err := u.replaceBlocks.Execute(ctx, ReplacePageBlocksInput{
			WorkspaceID:  in.WorkspaceID,
			PageID:       in.PageID,
			Doc:          s.Doc,
			EditorUserID: in.ResolverUserID,
			ForceVersion: true,
		}); err != nil {
			return err
		}
		resolved = s
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resolved, nil
}

// RejectPageSuggestionUseCase は提案を却下する。本文は一切触らない。
type RejectPageSuggestionUseCase struct {
	suggestions repository.PageSuggestionRepository
}

func NewRejectPageSuggestionUseCase(suggestions repository.PageSuggestionRepository) *RejectPageSuggestionUseCase {
	return &RejectPageSuggestionUseCase{suggestions: suggestions}
}

type RejectSuggestionInput struct {
	WorkspaceID    string
	PageID         string
	SuggestionID   string
	ResolverUserID uint64
}

func (u *RejectPageSuggestionUseCase) Execute(ctx context.Context, in RejectSuggestionInput) (*domain.PageSuggestion, error) {
	return u.suggestions.Resolve(
		ctx, in.WorkspaceID, in.PageID, in.SuggestionID, domain.PageSuggestionStatusRejected, in.ResolverUserID, time.Now(),
	)
}
