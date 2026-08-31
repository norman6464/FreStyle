package repository

import (
	"context"
	"errors"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// ErrWorkspaceNotFound は対象ワークスペースが存在しないときに返す。
var ErrWorkspaceNotFound = errors.New("workspace not found")

// ErrWorkspaceHasMembers は所属している人がいるワークスペースを消そうとしたときに返す。
// そこには全員のノートが入るので、1 人の操作で消せてよいはずがない。
var ErrWorkspaceHasMembers = errors.New("workspace still has members")

// ErrSpaceNotFound は対象スペースが存在しない（または別ワークスペースのもの）ときに返す。
var ErrSpaceNotFound = errors.New("space not found")

// ErrWorkspaceSlugTaken は作成しようとした slug が既に使われているときに返す。
// slug はグローバルに一意なので、他テナントが使っている場合もこれになる
// （どのテナントが使っているかは返さない）。
var ErrWorkspaceSlugTaken = errors.New("workspace slug is already taken")

// ErrSpaceKeyTaken は作成しようとした key が同じワークスペースで既に使われているときに返す。
var ErrSpaceKeyTaken = errors.New("space key is already taken")

// ErrPersonalWorkspaceAlreadyExists は、その人の個人ワークスペースを新規作成しようとした瞬間に
// 別のリクエストが先に作り終えていたときに返す（uq_workspaces_personal_owner の競合）。
// サインアップの二重送信・同時実行で起き得る。呼び出し側は失敗として扱わず、
// FindPersonalWorkspaceByOwner で先に作られた方を引き直す。
var ErrPersonalWorkspaceAlreadyExists = errors.New("personal workspace already exists for this user")

// ErrPageNotFound は対象ページが存在しない（または別ワークスペースのもの）ときに返す。
// テナント越えのアクセスは「無い」と同じ扱いにする（存在の有無自体を漏らさない）。
var ErrPageNotFound = errors.New("page not found")

// ErrPageMoveVoidsSpaceRestriction は、移動するサブツリーに「スペース全員」宛ての例外が
// 残っている状態で、そのスペースの外へ移そうとしたときに返す。
//
// space_all の主体は「そのスペースの全員」を意味するため、ページが別スペースへ移ると
// 行は残ったまま評価の対象から外れる。deny なら誰も外れなくなって（＝ 開いて）しまい、
// allow なら誰も載っていない許可リストになって全員が締め出される。どちらも
// 「権限設定画面に見えているものと実効が違う」状態で、移動した本人にも気づけない。
// 例外を先に整理してから移す運用に倒し、黙って権限が変わる経路を塞ぐ。
var ErrPageMoveVoidsSpaceRestriction = errors.New("page move would void a space-wide restriction")

// ErrPageSnapshotNotFound は対象ページの snapshot がまだ無いときに返す。
// snapshot は派生データなので、呼び出し側はこれを受けて blocks から組み立てる。
var ErrPageSnapshotNotFound = errors.New("page snapshot not found")

// BlockWrite は ReplacePageBlocks に渡す 1 ブロック行。
//
// ID を持たないのは採番（UUIDv7）が repository の責務のため。親子関係は保存前に ID が
// 決まらないので、同じスライス内の親の添字（ParentIndex）で表す。文書順に並べる
// （親は必ず子より前 = ParentIndex < 自分の添字）ことが前提で、実装側はこれを検証する。
type BlockWrite struct {
	// ParentIndex は親ブロックの添字。-1 はページ直下（トップレベル）。
	ParentIndex int
	// Position は兄弟内の並び順（分数インデックス）。
	Position string
	// Type は ProseMirror のノード名。
	Type domain.BlockType
	// Attrs は ProseMirror の attrs（JSON object 文字列。属性が無ければ "{}"）。
	Attrs string
	// Inline は葉ノードのインライン内容（JSON array 文字列）。容器ノードは nil。
	Inline *string
}

// KnowledgeBaseRepository はノート（workspaces / spaces / pages / blocks /
// page_paths / page_snapshots）へのアクセスを提供する。
//
// 1 boundary = 1 fat interface（§2.6）。ページ・ブロック・closure・snapshot は
// 「ページ作成 = pages + page_paths」「本文保存 = blocks + snapshot」のように
// 複数テーブルを 1 トランザクションで書く操作が中心で、境界を分けると
// トランザクション境界が interface をまたいでしまうため 1 つにまとめる。
// トランザクションは実装内で完結させ、usecase に *sql.Tx を漏らさない。
type KnowledgeBaseRepository interface {
	// FindPageByIDAcrossWorkspaces はページを ID だけで引く（/p/{pageId} の解決用）。
	// このリポジトリで唯一テナントを確定せずに読む口で、呼び出し側は結果を応答に使う前に
	// **必ずその workspace の権限判定を通す**。無ければ ErrPageNotFound。
	FindPageByIDAcrossWorkspaces(ctx context.Context, pageID string) (*domain.Page, error)
	// ListAncestorPageIDs はページの祖先 ID を根から順に返す（自分自身は含まない）。
	// パンくず用の骨組み。題名・可視性は返さない — 可視の判定は権限側の口が持つ。
	// ページが無い・根ページなら空（エラーにしない。実在の確認は呼び出し側が済ませている）。
	ListAncestorPageIDs(ctx context.Context, workspaceID, pageID string) ([]string, error)
	// DeleteWorkspace はワークスペースを配下ごと消す（FK の CASCADE で連なる）。
	// **所属している人がいるワークスペースは消さない** — その場合 ErrWorkspaceHasMembers。
	// 対象が無ければ ErrWorkspaceNotFound（消えたことにしない — 押した人に結果を返す）。
	DeleteWorkspace(ctx context.Context, workspaceID string) error

	// FindWorkspaceByID はワークスペースを 1 件引く。無ければ ErrWorkspaceNotFound。
	FindWorkspaceByID(ctx context.Context, workspaceID string) (*domain.Workspace, error)
	// FindWorkspaceBySlug は URL に出る slug からワークスペースを引く。無ければ ErrWorkspaceNotFound。
	FindWorkspaceBySlug(ctx context.Context, slug string) (*domain.Workspace, error)
	// FindPersonalWorkspaceByOwner はそのユーザーの個人ワークスペースを引く。無ければ
	// ErrWorkspaceNotFound（uq_workspaces_personal_owner が 1 人 1 つを守るので、
	// 見つかれば必ず 1 件）。サインアップの「作る前に既に在るか見る」に使う。
	FindPersonalWorkspaceByOwner(ctx context.Context, userID uint64) (*domain.Workspace, error)
	// FindSpace はスペースを 1 件引く。無い・別ワークスペースなら ErrSpaceNotFound。
	FindSpace(ctx context.Context, workspaceID, spaceID string) (*domain.Space, error)
	// UpdateSpaceName はスペースの表示名だけを変える（key は URL・権限の参照に使うので不変）。
	// 0 件更新（無い・別ワークスペース）は ErrSpaceNotFound。
	UpdateSpaceName(ctx context.Context, workspaceID, spaceID, name string) error
	// CreateSpace はスペースを作成する。ID は UUIDv7 を採番して space.ID に反映し、
	// 呼び出し後の space は DB で確定した行（created_at 等）で上書きされる。
	// key が同じワークスペースで使用済みなら ErrSpaceKeyTaken。
	//
	// 「全員」の主体（kind='space_all'）はここでは作らない。あれは grant を張るときに
	// 初めて要る主体で、EnsureSpaceEveryonePrincipal が必要な時点で作る
	// （どのスペースにも必ず 1 つ、という不変条件を持たせると、消えたときに
	// 直す責任の所在が分からなくなる）。
	CreateSpace(ctx context.Context, space *domain.Space) error
	// FindPage はページを 1 件引く（アーカイブ済みも返す）。無い・別ワークスペースなら ErrPageNotFound。
	FindPage(ctx context.Context, workspaceID, pageID string) (*domain.Page, error)
	// ListActivePagesBySpace はスペース配下の現役ページ全件を position 順で返す（ツリー構築用）。
	ListActivePagesBySpace(ctx context.Context, workspaceID, spaceID string) ([]domain.Page, error)
	// LastActiveSiblingPosition は兄弟（parentID が nil ならスペース直下）の末尾 position を返す。
	// 兄弟がいなければ空文字（fracindex.Between の「端」表現）。
	LastActiveSiblingPosition(ctx context.Context, workspaceID, spaceID string, parentID *string) (string, error)
	// SiblingPositionsAround は「ある兄弟の隣に入れる」ための前後の並び順キーを返す。
	//
	// ドラッグで落とした位置を表すのに使う。**クライアントは並び順のキーを持たない**ので
	// （応答に入れていない。整数部が兄弟の通し番号になるため、飛びから伏せた枚数が読める）、
	// 位置は「どの兄弟の隣か」をページの ID で受け取り、キーの計算はこちら側で閉じる。
	//
	// found が false なら anchorPageID はその親の現役の子ではない（不在・別の親・
	// 別スペース・アーカイブ済みを区別しない）。前後の端が無いことは空文字で表す
	// （fracindex.Between の約束と同じ）。
	//
	// movingPageID は必ず除く。動かす当人がまだその並びに居るので、除かないと
	// 自分自身との中間値を計算することになる。空文字なら「除くものが無い」
	// （まだ並びに居ないものを差し込むとき用）。
	SiblingPositionsAround(
		ctx context.Context, workspaceID, spaceID string, parentID *string, anchorPageID, movingPageID string,
	) (found bool, prev, anchorPos, next string, err error)
	// HasActiveSiblingPosition は excludePageID 以外の現役の兄弟が position を使用中かを返す
	// （アーカイブ復帰時の衝突検出用）。
	HasActiveSiblingPosition(ctx context.Context, workspaceID, spaceID string, parentID *string, position, excludePageID string) (bool, error)
	// HasDescendant は candidateID が pageID の子孫（自分自身を含む）かを返す（移動の循環検出用）。
	HasDescendant(ctx context.Context, workspaceID, pageID, candidateID string) (bool, error)
	// CreatePage はページを作成する。ID は UUIDv7 を採番して page.ID に反映し、
	// closure（自分自身 depth=0 + 祖先の組）も同一トランザクションで張る。
	// 呼び出し後の page は DB で確定した行（created_at 等）で上書きされる。
	CreatePage(ctx context.Context, page *domain.Page) error
	// UpdatePageTitle はタイトルを変更し、更新後の行を返す。対象が無ければ ErrPageNotFound。
	UpdatePageTitle(ctx context.Context, workspaceID, pageID, title string) (*domain.Page, error)
	// MovePage はページを newParentID（nil はスペース直下）の末尾へ移す。
	// pages の付け替え・スペースが変わる場合のサブツリー space_id 更新・closure の
	// 付け替えを 1 トランザクションで行う。対象が無ければ ErrPageNotFound。
	// スペースをまたぐ移動で、サブツリーに移動先以外のスペースの「全員」宛て例外が
	// 残っている場合は ErrPageMoveVoidsSpaceRestriction を返して移動しない。
	MovePage(ctx context.Context, workspaceID, pageID string, newParentID *string, newSpaceID, newPosition string) error
	// DeletePageSubtree はページを子孫ごと物理削除する（closure・blocks・snapshot も
	// CASCADE で消える）。対象が無ければ ErrPageNotFound。アーカイブと違い戻せないため、
	// 子孫全員の編集権限の確認は呼び出し側の入口が行う。
	DeletePageSubtree(ctx context.Context, workspaceID, pageID string) error
	// ArchivePageSubtree はページとその子孫のうち現役の行に archived_at を設定する。
	// 既にアーカイブ済みの行は元の archived_at を保つ（触らない）。
	ArchivePageSubtree(ctx context.Context, workspaceID, pageID string) error
	// UnarchivePageSubtree はサブツリーのうち archivedSince 以降にアーカイブされた行
	// （= 根と同時にアーカイブされた一括分）を現役へ戻す。newRootPosition が非 nil なら
	// 解除前に根の position を振り直す（現役の兄弟との衝突回避）。同一トランザクションで行う。
	UnarchivePageSubtree(ctx context.Context, workspaceID, pageID string, archivedSince time.Time, newRootPosition *string) error
	// ListBlocksByPage はページの全ブロックを position 順で返す（doc への組み立て用）。
	ListBlocksByPage(ctx context.Context, workspaceID, pageID string) ([]domain.Block, error)
	// ReplacePageBlocks はページの全ブロックを blocks で置き換え、snapshot を snapshotDoc で
	// 焼き直す（全消し全入れ + UPSERT を 1 トランザクションで）。対象ページが無ければ ErrPageNotFound。
	ReplacePageBlocks(ctx context.Context, workspaceID, pageID string, blocks []BlockWrite, snapshotDoc string) error
	// GetPageSnapshot はページの snapshot を返す。無ければ ErrPageSnapshotNotFound、
	// ページ自体が別ワークスペースなら ErrPageSnapshotNotFound と同じ「無い」に落ちる。
	GetPageSnapshot(ctx context.Context, workspaceID, pageID string) (*domain.PageSnapshot, error)
}
