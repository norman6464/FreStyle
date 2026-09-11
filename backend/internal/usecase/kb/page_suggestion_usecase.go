package kb

import (
	"context"
	"errors"
	"time"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// 提案（page_suggestions）は kb パッケージ直下に置く。comment のような独立パッケージには
// しない — 提案はページ本文そのものの書き込み経路の分岐であり、認可も既存の
// CanComment / CanEdit だけで足りる（page_version_usecase.go の doc と同じ判断）。

// ErrTooManyOpenSuggestions は未解決（open）の提案が上限に達しているときに CreateSuggestionUseCase
// が返す。commenter は編集権限を持たないため、この上限に引っかかった側にできることは
// 既存の提案が解決されるのを待つか、編集者に却下・採用を促すことだけ。
//
// 上限は 2 段構え:
//   - maxOpenSuggestionsPerAuthorPerPage: 投稿者 1 人が同じページへ積める数。
//     1 人が自動化で大量投稿する攻撃を、他の投稿者を巻き込まずに抑える。
//   - maxOpenSuggestionsPerPage: ページ全体で溜められる数。複数アカウントに分散されても、
//     編集者がレビューする一覧そのものが際限なく膨らまないようにする（frontend の
//     computeSuggestionDiff / KbSuggestionDiffView の負荷上限とは独立の防御）。
var ErrTooManyOpenSuggestions = errors.New("too many open page suggestions")

const (
	maxOpenSuggestionsPerAuthorPerPage = 20
	maxOpenSuggestionsPerPage          = 100
)

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
	// 上限判定は入力の妥当性検証より前に行ってよい — doc のパースが軽い操作だとしても、
	// 既に溜まっている提案の数はその投稿の中身と無関係に決まるので、先に弾いた方が無駄がない。
	byAuthor, err := u.suggestions.CountOpenByAuthor(ctx, in.WorkspaceID, in.PageID, in.AuthorUserID)
	if err != nil {
		return nil, err
	}
	if byAuthor >= maxOpenSuggestionsPerAuthorPerPage {
		return nil, ErrTooManyOpenSuggestions
	}
	total, err := u.suggestions.CountOpen(ctx, in.WorkspaceID, in.PageID)
	if err != nil {
		return nil, err
	}
	if total >= maxOpenSuggestionsPerPage {
		return nil, ErrTooManyOpenSuggestions
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

// defaultOpenSuggestionsLimit / maxOpenSuggestionsLimit は ListOpen の SQL LIMIT を決める。
// 既定値を maxOpenSuggestionsPerPage に揃えているのは、書き込み側の上限が効いている限り
// 通常の一覧取得が 1 ページで全 open 提案を見せられるようにするため（呼び出し元は今のところ
// 頁送りの UI を持たない）。呼び出し元が明示的に Limit を指定すれば、この既定より絞れる
// （上振れは maxOpenSuggestionsLimit で頭打ちにし、書き込み側の上限を超えて要求されても
// SQL 側の負荷を増やさない）。
const (
	defaultOpenSuggestionsLimit = maxOpenSuggestionsPerPage
	maxOpenSuggestionsLimit     = 200
)

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
	// Limit は返す最大件数。0 以下なら defaultOpenSuggestionsLimit。maxOpenSuggestionsLimit
	// を超える値は切り詰める（SearchViewablePagesUseCase.Limit と同じ挟み方）。
	Limit int
}

func (u *ListOpenPageSuggestionsUseCase) Execute(ctx context.Context, in ListOpenPageSuggestionsInput) ([]domain.PageSuggestion, error) {
	limit := in.Limit
	if limit <= 0 {
		limit = defaultOpenSuggestionsLimit
	}
	if limit > maxOpenSuggestionsLimit {
		limit = maxOpenSuggestionsLimit
	}
	return u.suggestions.ListOpen(ctx, in.WorkspaceID, in.PageID, limit)
}

// AcceptPageSuggestionUseCase は提案を採用する — 提案の doc を通常の保存経路
// （ReplacePageBlocksUseCase）へそのまま渡し、本文へ反映しつつ版を 1 つ切る。
//
// 提案の解決（Resolve）と本文の書き換え（replaceBlocks.Execute）は 1 つのトランザクションに
// 入れる。Resolve が先に 'accepted' にした後で replaceBlocks.Execute が失敗したら、
// 提案が'accepted'のまま本文だけ古い、という中間状態を作らないため（両方ロールバックする）。
type AcceptPageSuggestionUseCase struct {
	suggestions   repository.PageSuggestionRepository
	versionRepo   repository.PageVersionRepository
	replaceBlocks *ReplacePageBlocksUseCase
	txManager     repository.TxManager
}

func NewAcceptPageSuggestionUseCase(
	suggestions repository.PageSuggestionRepository, versionRepo repository.PageVersionRepository,
	replaceBlocks *ReplacePageBlocksUseCase, txManager repository.TxManager,
) *AcceptPageSuggestionUseCase {
	return &AcceptPageSuggestionUseCase{
		suggestions: suggestions, versionRepo: versionRepo, replaceBlocks: replaceBlocks, txManager: txManager,
	}
}

type AcceptSuggestionInput struct {
	WorkspaceID    string
	PageID         string
	SuggestionID   string
	ResolverUserID uint64
}

// suggestionIsStale は、提案した時点の版（baseSeq）より後にそのページが編集されたかを判定する。
// 差分は baseSeq が指す版を基準に計算されているため、その後の編集を知らない。ずれていれば
// このまま採用すると後の編集を黙って本文ごと上書きしてしまうので、採用者に一度突き返す。
func suggestionIsStale(baseSeq *int64, latest *domain.PageVersion) bool {
	if baseSeq == nil {
		// 提案した時点で版が 1 つも無かった。以後に 1 つでも版ができていれば、それは
		// 提案作成後の編集（本文保存 / 版を残す / 復元のいずれか）に他ならない。
		return latest != nil
	}
	return latest == nil || latest.Seq != *baseSeq
}

func (u *AcceptPageSuggestionUseCase) Execute(ctx context.Context, in AcceptSuggestionInput) (*domain.PageSuggestion, error) {
	var resolved *domain.PageSuggestion
	err := u.txManager.DoInTx(ctx, func(ctx context.Context) error {
		s, err := u.suggestions.Get(ctx, in.WorkspaceID, in.PageID, in.SuggestionID)
		if err != nil {
			return err
		}
		if s.Status != domain.PageSuggestionStatusOpen {
			return domain.ErrPageSuggestionAlreadyResolved
		}
		latest, err := u.versionRepo.GetLatestVersion(ctx, in.WorkspaceID, in.PageID)
		if err != nil {
			return err
		}
		if suggestionIsStale(s.BaseSeq, latest) {
			return domain.ErrPageSuggestionStale
		}
		resolvedSuggestion, err := u.suggestions.Resolve(
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
			Doc:          resolvedSuggestion.Doc,
			EditorUserID: in.ResolverUserID,
			ForceVersion: true,
		}); err != nil {
			return err
		}
		resolved = resolvedSuggestion
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
