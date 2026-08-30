package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
)

// リッチテキスト文書のユースケース向けエラー（handler が HTTP ステータスへ対応づける）。
// repository の内部エラーはここで usecase 層のセンチネルへ翻訳し、handler が repository を知らずに済むようにする。
var (
	// ErrRichDocumentNotFound は文書が存在しない／閲覧権が無い（存在を漏らさないため統一）。→ 404。
	ErrRichDocumentNotFound = errors.New("rich document not found")
	// ErrRichDocumentConflict は楽観ロックの版不一致。→ 409。
	ErrRichDocumentConflict = errors.New("rich document revision conflict")
	// ErrRichDocumentForbidden は他人の文書を更新/削除しようとした。→ 403。
	ErrRichDocumentForbidden = errors.New("forbidden")
	// ErrRichDocumentInvalid は入力（doc / title / kind）が不正。→ 400。
	ErrRichDocumentInvalid = errors.New("invalid rich document")
)

// maxDocBytes は doc(jsonb) のサイズ上限。画像は S3 に置き本文に埋めないため、この上限で編集を軽く保つ。
const maxDocBytes = 1 << 20 // 1 MiB

const maxTitleLen = 200

// validateDoc は保存する doc が「tiptap のドキュメント JSON（object かつ type='doc'）」で、
// サイズ上限内であることを検証する。PostgreSQL の jsonb は U+0000（NUL）を格納できないため、
// アプリ側でも弾いて DB 挿入時の 500 を 400 に前倒しする（DB の CHECK と併せた多層の壁）。
func validateDoc(doc string) error {
	if len(doc) == 0 {
		return fmt.Errorf("%w: doc is required", ErrRichDocumentInvalid)
	}
	if len(doc) > maxDocBytes {
		return fmt.Errorf("%w: doc exceeds %d bytes", ErrRichDocumentInvalid, maxDocBytes)
	}
	// 生の NUL バイト（0x00）は jsonb に入らない。encoding/json も必ず弾くが、明快な 400 を早期に返す。
	// 生バイトは正当な入力になり得ないので誤検知は起きない。
	if strings.ContainsRune(doc, 0) {
		return fmt.Errorf("%w: doc must not contain NUL (U+0000)", ErrRichDocumentInvalid)
	}
	var v any
	if err := json.Unmarshal([]byte(doc), &v); err != nil {
		return fmt.Errorf("%w: doc must be a JSON object", ErrRichDocumentInvalid)
	}
	// エスケープ表記 \u0000 はデコード後に実 NUL ルーンになる。デコード済みの全文字列値を走査して
	// 実際の NUL のみ弾く（生 JSON への部分一致だと、正当な文字列値 "\\u0000"（6 文字リテラル）を
	// 誤って弾いてしまうため）。
	if jsonHasNULRune(v) {
		return fmt.Errorf("%w: doc must not contain NUL (U+0000)", ErrRichDocumentInvalid)
	}
	m, ok := v.(map[string]any)
	if !ok {
		return fmt.Errorf("%w: doc must be a JSON object", ErrRichDocumentInvalid)
	}
	if typ, _ := m["type"].(string); typ != "doc" {
		return fmt.Errorf("%w: doc.type must be \"doc\"", ErrRichDocumentInvalid)
	}
	return nil
}

// jsonHasNULRune は json.Unmarshal 済みの値ツリー内の文字列（オブジェクトのキーを含む）に
// 実 NUL ルーンがあるかを再帰判定する。エスケープ \u0000 はここで実 NUL にデコードされ捕捉され、
// リテラル "\\u0000"（文字列としての 6 文字）は素通りする。
func jsonHasNULRune(v any) bool {
	switch t := v.(type) {
	case string:
		return strings.ContainsRune(t, 0)
	case []any:
		for _, e := range t {
			if jsonHasNULRune(e) {
				return true
			}
		}
	case map[string]any:
		for k, e := range t {
			if strings.ContainsRune(k, 0) || jsonHasNULRune(e) {
				return true
			}
		}
	}
	return false
}

// containsNUL はデコード済みの文字列（title など）に実 NUL ルーンが含まれるかを返す。jsonb は格納できない。
func containsNUL(s string) bool {
	return strings.ContainsRune(s, 0)
}

func validateTitle(title string) error {
	if title == "" {
		return fmt.Errorf("%w: title is required", ErrRichDocumentInvalid)
	}
	if len([]rune(title)) > maxTitleLen {
		return fmt.Errorf("%w: title too long", ErrRichDocumentInvalid)
	}
	if containsNUL(title) {
		return fmt.Errorf("%w: title must not contain NUL", ErrRichDocumentInvalid)
	}
	return nil
}

// GetRichDocumentUseCase は 1 文書を取得する（閲覧認可込み）。
type GetRichDocumentUseCase struct {
	repo repository.RichDocumentRepository
}

func NewGetRichDocumentUseCase(r repository.RichDocumentRepository) *GetRichDocumentUseCase {
	return &GetRichDocumentUseCase{repo: r}
}

// Execute は id の文書を返す。閲覧可否の規則は domain.RichDocument.CanBeReadBy が唯一持ち、
// ここでは条件を写経しない（SQL 側にも書かない）。閲覧できない場合は存在を漏らさないため
// ErrRichDocumentNotFound を返す。viewerID=0 は未認証、viewerWorkspace は閲覧者の所属ワークスペース。
func (u *GetRichDocumentUseCase) Execute(ctx context.Context, id string, viewerID uint64, viewerWorkspace domain.WorkspaceRef) (*domain.RichDocument, error) {
	doc, err := u.repo.FindByID(ctx, id)
	if err != nil {
		return nil, translateRepoErr(err)
	}
	if !doc.CanBeReadBy(viewerID, viewerWorkspace) {
		return nil, ErrRichDocumentNotFound
	}
	return doc, nil
}

// CreateRichDocumentUseCase は新規文書を作成する。
type CreateRichDocumentUseCase struct {
	repo repository.RichDocumentRepository
}

func NewCreateRichDocumentUseCase(r repository.RichDocumentRepository) *CreateRichDocumentUseCase {
	return &CreateRichDocumentUseCase{repo: r}
}

type CreateRichDocumentInput struct {
	OwnerID       uint64
	CompanyID     *uint64
	WorkspaceID   *string
	Kind          domain.DocumentKind
	Title         string
	Doc           string
	IsPublic      bool
	SchemaVersion int
}

func (u *CreateRichDocumentUseCase) Execute(ctx context.Context, in CreateRichDocumentInput) (*domain.RichDocument, error) {
	if in.OwnerID == 0 {
		return nil, fmt.Errorf("%w: ownerID is required", ErrRichDocumentInvalid)
	}
	if !in.Kind.Valid() {
		return nil, fmt.Errorf("%w: unknown kind %q", ErrRichDocumentInvalid, in.Kind)
	}
	if err := validateTitle(in.Title); err != nil {
		return nil, err
	}
	if err := validateDoc(in.Doc); err != nil {
		return nil, err
	}
	sv := in.SchemaVersion
	if sv <= 0 {
		sv = 1
	}
	doc := &domain.RichDocument{
		OwnerID:       in.OwnerID,
		WorkspaceID:   in.WorkspaceID,
		Kind:          in.Kind,
		Title:         in.Title,
		IsPublic:      in.IsPublic,
		SchemaVersion: sv,
		Doc:           in.Doc,
		Revision:      1,
	}
	if err := u.repo.Create(ctx, doc); err != nil {
		return nil, translateRepoErr(err)
	}
	return doc, nil
}

// UpdateRichDocumentUseCase は文書を更新する（所有者検証＋楽観ロック）。
type UpdateRichDocumentUseCase struct {
	repo repository.RichDocumentRepository
}

func NewUpdateRichDocumentUseCase(r repository.RichDocumentRepository) *UpdateRichDocumentUseCase {
	return &UpdateRichDocumentUseCase{repo: r}
}

type UpdateRichDocumentInput struct {
	ID            string
	ActorID       uint64
	Title         string
	Doc           string
	IsPublic      bool
	SchemaVersion int
	Revision      int
}

func (u *UpdateRichDocumentUseCase) Execute(ctx context.Context, in UpdateRichDocumentInput) (*domain.RichDocument, error) {
	// revision は 1 始まりで増分のみ。負数/0 は決して一致せず楽観ロックの版不一致（409）に化けるため、
	// 不正入力として 400 で早期に弾く（handler の binding:"required" は 0 は弾くが負数は通る）。
	if in.Revision < 1 {
		return nil, fmt.Errorf("%w: revision must be positive", ErrRichDocumentInvalid)
	}
	existing, err := u.repo.FindByID(ctx, in.ID)
	if err != nil {
		return nil, translateRepoErr(err)
	}
	if existing.OwnerID != in.ActorID {
		// 非所有者には存在を漏らさない（Get / Delete と揃えて 404 にする）。
		return nil, ErrRichDocumentNotFound
	}
	if err := validateTitle(in.Title); err != nil {
		return nil, err
	}
	if err := validateDoc(in.Doc); err != nil {
		return nil, err
	}
	sv := in.SchemaVersion
	if sv <= 0 {
		sv = existing.SchemaVersion
	}
	doc := &domain.RichDocument{
		ID:            in.ID,
		Title:         in.Title,
		IsPublic:      in.IsPublic,
		SchemaVersion: sv,
		Doc:           in.Doc,
	}
	if err := u.repo.UpdateWithRevision(ctx, doc, in.Revision); err != nil {
		return nil, translateRepoErr(err)
	}
	return doc, nil
}

// DeleteRichDocumentUseCase は文書を論理削除する（所有者のみ）。
type DeleteRichDocumentUseCase struct {
	repo repository.RichDocumentRepository
}

func NewDeleteRichDocumentUseCase(r repository.RichDocumentRepository) *DeleteRichDocumentUseCase {
	return &DeleteRichDocumentUseCase{repo: r}
}

func (u *DeleteRichDocumentUseCase) Execute(ctx context.Context, id string, actorID uint64) error {
	if actorID == 0 {
		return ErrRichDocumentForbidden
	}
	return translateRepoErr(u.repo.SoftDelete(ctx, id, actorID))
}

// ListRichDocumentsUseCase は owner 本人の文書一覧を返す（軽量サマリ）。
type ListRichDocumentsUseCase struct {
	repo repository.RichDocumentRepository
}

func NewListRichDocumentsUseCase(r repository.RichDocumentRepository) *ListRichDocumentsUseCase {
	return &ListRichDocumentsUseCase{repo: r}
}

type ListRichDocumentsInput struct {
	// OwnerID は一覧の対象であり閲覧者本人でもある（この経路は owner スコープ固定）。
	OwnerID uint64
	// ViewerWorkspace は閲覧者の所属ワークスペース。可視性の判定を domain へ渡すために持つ。
	ViewerWorkspace domain.WorkspaceRef
	// Kind が空なら全 kind。指定するなら既知の kind であること。
	Kind domain.DocumentKind
}

func (u *ListRichDocumentsUseCase) Execute(ctx context.Context, in ListRichDocumentsInput) ([]domain.RichDocument, error) {
	if in.OwnerID == 0 {
		return nil, fmt.Errorf("%w: ownerID is required", ErrRichDocumentInvalid)
	}
	if in.Kind != "" && !in.Kind.Valid() {
		return nil, fmt.Errorf("%w: unknown kind %q", ErrRichDocumentInvalid, in.Kind)
	}
	rows, err := u.repo.ListByOwner(ctx, in.OwnerID, in.Kind)
	if err != nil {
		return nil, translateRepoErr(err)
	}
	// ListByOwner は owner スコープなので、返る行は必ず「閲覧者＝所有者」を満たす。
	// それでも同じ domain の規則を通しておく。将来この一覧が所有者以外の文書を含むよう
	// 広がったときに、可視性の判断がここへ写経されるのではなく CanBeReadBy に残るようにするため。
	visible := rows[:0]
	for _, row := range rows {
		if row.CanBeReadBy(in.OwnerID, in.ViewerWorkspace) {
			visible = append(visible, row)
		}
	}
	return visible, nil
}

// translateRepoErr は repository 層のセンチネルを usecase 層のセンチネルへ翻訳する。
func translateRepoErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, repository.ErrRichDocumentNotFound):
		return ErrRichDocumentNotFound
	case errors.Is(err, repository.ErrRichDocumentConflict):
		return ErrRichDocumentConflict
	case errors.Is(err, repository.ErrRichDocumentInvalidData):
		return ErrRichDocumentInvalid
	default:
		return err
	}
}
