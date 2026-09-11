package kb

import (
	"context"
	"fmt"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// ページの雛形（page_templates）は kb パッケージ直下に置く。comment のような独立パッケージには
// しない — 雛形は新しい権限軸を持たず、認可はワークスペース / ページの CanEdit・CanView
// だけで足りる（page_version_usecase.go の doc と同じ判断）。

// CreateTemplateFromPageUseCase は既存ページの「今の」本文を雛形として保存する
// （「雛形として保存」）。
type CreateTemplateFromPageUseCase struct {
	kbRepo     repository.KnowledgeBaseRepository
	templates  repository.PageTemplateRepository
	checkSpace *CheckSpacePermissionUseCase
}

func NewCreateTemplateFromPageUseCase(
	kbRepo repository.KnowledgeBaseRepository, templates repository.PageTemplateRepository,
	checkSpace *CheckSpacePermissionUseCase,
) *CreateTemplateFromPageUseCase {
	return &CreateTemplateFromPageUseCase{kbRepo: kbRepo, templates: templates, checkSpace: checkSpace}
}

type CreateTemplateFromPageInput struct {
	WorkspaceID string
	PageID      string
	// SpaceID が nil ならワークスペース全体で見える雛形になる。非 nil ならそのスペース限定
	// （実在確認に加え、呼び出し者がそのスペースを閲覧できることも確かめる。
	// どちらも満たさなければ repository.ErrSpaceNotFound — 「見えない」と「無い」を
	// 区別しない既存の方針どおり）。
	SpaceID      *string
	Name         string
	AuthorUserID uint64
}

func (u *CreateTemplateFromPageUseCase) Execute(ctx context.Context, in CreateTemplateFromPageInput) (*domain.PageTemplate, error) {
	// 名前の検証は repository を呼ぶ前に済ませる（SetPageIconUseCase と同じ理由 —
	// 不正な値で先に読みに行くと、「値は捨てられたが読みには行った」という中途半端な
	// 副作用だけが残る）。
	name, err := domain.ValidateTemplateName(in.Name)
	if err != nil {
		return nil, err
	}
	page, err := u.kbRepo.FindPage(ctx, in.WorkspaceID, in.PageID)
	if err != nil {
		return nil, err
	}
	// GetPageUseCase.Execute と全く同じフォールバック手順（snapshot 優先 → 無ければ
	// blocks から組み立て）を currentPageDoc（page_version_usecase.go）へ委ねる。
	doc, err := currentPageDoc(ctx, u.kbRepo, in.WorkspaceID, in.PageID)
	if err != nil {
		return nil, err
	}
	stripped, err := stripPageRefAndImageNodesForTemplate(doc)
	if err != nil {
		return nil, err
	}
	if in.SpaceID != nil {
		if _, err := u.kbRepo.FindSpace(ctx, in.WorkspaceID, *in.SpaceID); err != nil {
			return nil, err
		}
		// 本文を読んだ・書けたページ（in.PageID）と、雛形をひも付けようとしている先の
		// スペース（in.SpaceID）は別物になり得る。クライアントが SpaceID を自由に選べる
		// リクエストなので、実在確認だけでは「見たこともない非公開スペースへ、
		// 自分が読める別ページの本文を紐付ける」ことを止められない。
		perm, err := u.checkSpace.Execute(ctx, CheckSpacePermissionInput{
			WorkspaceID: in.WorkspaceID, SpaceID: *in.SpaceID, UserID: in.AuthorUserID,
		})
		if err != nil {
			return nil, err
		}
		if !perm.CanView {
			return nil, repository.ErrSpaceNotFound
		}
	}
	tpl := &domain.PageTemplate{
		WorkspaceID: in.WorkspaceID,
		SpaceID:     in.SpaceID,
		Name:        name,
		// 元ページのアイコンをそのままコピーする。特別な検証・加工は要らない
		// （既に domain.PageIcon.Valid() を満たした状態で page に保存されている）。
		Icon:            page.Icon,
		Doc:             stripped,
		CreatedByUserID: in.AuthorUserID,
	}
	if err := u.templates.Create(ctx, tpl); err != nil {
		return nil, err
	}
	return tpl, nil
}

// ListPageTemplatesUseCase はワークスペース（または特定のスペース）の雛形一覧を返す。
type ListPageTemplatesUseCase struct {
	templates  repository.PageTemplateRepository
	checkSpace *CheckSpacePermissionUseCase
}

func NewListPageTemplatesUseCase(
	templates repository.PageTemplateRepository, checkSpace *CheckSpacePermissionUseCase,
) *ListPageTemplatesUseCase {
	return &ListPageTemplatesUseCase{templates: templates, checkSpace: checkSpace}
}

type ListPageTemplatesInput struct {
	WorkspaceID string
	SpaceID     *string
	UserID      uint64
}

func (u *ListPageTemplatesUseCase) Execute(ctx context.Context, in ListPageTemplatesInput) ([]domain.PageTemplate, error) {
	// SpaceID 指定は「そのスペースの一覧」を名乗る。repository.List はワークスペース全体向け
	// （space_id IS NULL）の行にこの spaceID の行を足して返す仕様なので、閲覧権限を
	// 確かめずに通すと、非公開スペースの雛形名・アイコンを spaceId さえ分かれば
	// 誰でも読めてしまう。
	if in.SpaceID != nil {
		perm, err := u.checkSpace.Execute(ctx, CheckSpacePermissionInput{
			WorkspaceID: in.WorkspaceID, SpaceID: *in.SpaceID, UserID: in.UserID,
		})
		if err != nil {
			return nil, err
		}
		if !perm.CanView {
			return nil, repository.ErrSpaceNotFound
		}
	}
	return u.templates.List(ctx, in.WorkspaceID, in.SpaceID)
}

// DeletePageTemplateUseCase は雛形を削除する。
type DeletePageTemplateUseCase struct {
	templates  repository.PageTemplateRepository
	checkSpace *CheckSpacePermissionUseCase
}

func NewDeletePageTemplateUseCase(
	templates repository.PageTemplateRepository, checkSpace *CheckSpacePermissionUseCase,
) *DeletePageTemplateUseCase {
	return &DeletePageTemplateUseCase{templates: templates, checkSpace: checkSpace}
}

type DeletePageTemplateInput struct {
	WorkspaceID string
	TemplateID  string
	UserID      uint64
}

func (u *DeletePageTemplateUseCase) Execute(ctx context.Context, in DeletePageTemplateInput) error {
	// handler はワークスペース全体への CanEdit しか確かめていない。雛形自身が非公開スペースに
	// ひも付いている（かつ呼び出し者がそのスペースを見られない）場合、それだけでは
	// 「そのスペースの雛形が存在すること」自体を実質的に確認・削除できてしまう
	// （List・CreateFromPage 側で塞いだ閲覧の穴と対になる書き込み側の穴）。
	tpl, err := u.templates.Get(ctx, in.WorkspaceID, in.TemplateID)
	if err != nil {
		return err
	}
	if tpl.SpaceID != nil {
		perm, err := u.checkSpace.Execute(ctx, CheckSpacePermissionInput{
			WorkspaceID: in.WorkspaceID, SpaceID: *tpl.SpaceID, UserID: in.UserID,
		})
		if err != nil {
			return err
		}
		if !perm.CanView {
			return domain.ErrPageTemplateNotFound
		}
	}
	return u.templates.Delete(ctx, in.WorkspaceID, in.TemplateID)
}

// CreatePageFromTemplateUseCase は雛形から新しいページを作る（「雛形から作る」）。
//
// CreatePageUseCase（空ページを作る）と ReplacePageBlocksUseCase（本文を書き込む）を
// この順で呼ぶだけの薄いオーケストレーション。どちらのシグネチャも変えない
// （RestorePageVersionUseCase が ReplacePageBlocksUseCase を注入されて呼ぶのと同じ形）。
type CreatePageFromTemplateUseCase struct {
	templates     repository.PageTemplateRepository
	checkSpace    *CheckSpacePermissionUseCase
	createPage    *CreatePageUseCase
	replaceBlocks *ReplacePageBlocksUseCase
	deletePage    *DeletePageUseCase
}

func NewCreatePageFromTemplateUseCase(
	templates repository.PageTemplateRepository,
	checkSpace *CheckSpacePermissionUseCase,
	createPage *CreatePageUseCase,
	replaceBlocks *ReplacePageBlocksUseCase,
	deletePage *DeletePageUseCase,
) *CreatePageFromTemplateUseCase {
	return &CreatePageFromTemplateUseCase{
		templates: templates, checkSpace: checkSpace,
		createPage: createPage, replaceBlocks: replaceBlocks, deletePage: deletePage,
	}
}

type CreatePageFromTemplateInput struct {
	WorkspaceID string
	SpaceID     string
	// ParentID が nil ならスペース直下（ルート）に作る（CreatePageUseCase.Input と同じ意味）。
	ParentID     *string
	TemplateID   string
	Title        string
	AuthorUserID uint64
}

func (u *CreatePageFromTemplateUseCase) Execute(ctx context.Context, in CreatePageFromTemplateInput) (*GetPageOutput, error) {
	tpl, err := u.templates.Get(ctx, in.WorkspaceID, in.TemplateID)
	if err != nil {
		return nil, err
	}
	// handler が確かめているのは「作成先の場所（親ページ or スペース直下）を編集できるか」
	// だけで、雛形そのものを見てよいかは一度も問われていない。雛形が非公開スペースに
	// ひも付いていれば、その本文（doc）はそのスペースの閲覧者だけに見せるべきもの
	// ——templateId さえ知っていれば作成先とは無関係に本文を抜き出せてしまう穴を塞ぐ。
	if tpl.SpaceID != nil {
		perm, err := u.checkSpace.Execute(ctx, CheckSpacePermissionInput{
			WorkspaceID: in.WorkspaceID, SpaceID: *tpl.SpaceID, UserID: in.AuthorUserID,
		})
		if err != nil {
			return nil, err
		}
		if !perm.CanView {
			return nil, domain.ErrPageTemplateNotFound
		}
	}
	// 同じ雛形から複数のページを作ると、剥がさないままだと blocks.id（グローバルに一意な PK）が
	// 衝突して 2 ページ目以降の保存が失敗する（regenerateBlockIDs の doc 参照）。
	doc, err := regenerateBlockIDs(tpl.Doc)
	if err != nil {
		return nil, err
	}
	page, err := u.createPage.Execute(ctx, CreatePageInput{
		WorkspaceID:     in.WorkspaceID,
		SpaceID:         in.SpaceID,
		ParentID:        in.ParentID,
		Title:           in.Title,
		CreatedByUserID: in.AuthorUserID,
	})
	if err != nil {
		return nil, err
	}
	snap, err := u.replaceBlocks.Execute(ctx, ReplacePageBlocksInput{
		WorkspaceID:  in.WorkspaceID,
		PageID:       page.ID,
		Doc:          doc,
		EditorUserID: in.AuthorUserID,
	})
	if err != nil {
		// 空ページの作成自体は成功しているので、後始末として削除を試みる。削除まで失敗した
		// 場合は、中途半端な空ページが残っていることをエラーメッセージに残す（呼び出し側の
		// ログから追えるように）。削除に成功すれば元のエラーをそのまま返す
		// （errors.Is による HTTP ステータスの判定を壊さないため）。
		if delErr := u.deletePage.Execute(ctx, DeletePageInput{WorkspaceID: in.WorkspaceID, PageID: page.ID}); delErr != nil {
			return nil, fmt.Errorf(
				"雛形からの本文書き込みに失敗し、作成済みの空ページ %s の後始末（削除）にも失敗しました（削除エラー: %w）: %w",
				page.ID, delErr, err,
			)
		}
		return nil, err
	}
	// 保存した本人が最終編集者になる（KnowledgeBasePageHandler.ReplaceContent と同じ扱い。
	// 再取得せずその場で反映する）。
	editorID := in.AuthorUserID
	page.LastEditedByUserID = &editorID
	builtAt := snap.BuiltAt
	return &GetPageOutput{Page: *page, Doc: snap.Doc, BuiltAt: &builtAt}, nil
}
