package kb

import (
	"context"
	"errors"
	"fmt"

	"github.com/norman6464/frestyle/backend/internal/domain"
	"github.com/norman6464/frestyle/backend/internal/usecase/repository"
)

// バージョン（page_versions）は kb パッケージ直下に置く。版はページ本文の履歴であり
// comment のような新しい権限軸を持たず、認可も CapabilityView / CapabilityEdit だけで足りる。

// currentPageDoc はページの「今の」本文（ProseMirror doc）を返す。GetPageUseCase.Execute と
// 全く同じフォールバック手順（snapshot 優先 → 無ければ blocks から組み立て）を踏む。
func currentPageDoc(ctx context.Context, repo repository.KnowledgeBaseRepository, workspaceID, pageID string) (string, error) {
	snap, err := repo.GetPageSnapshot(ctx, workspaceID, pageID)
	if err == nil {
		return snap.Doc, nil
	}
	if !errors.Is(err, repository.ErrPageSnapshotNotFound) {
		return "", err
	}
	blocks, err := repo.ListBlocksByPage(ctx, workspaceID, pageID)
	if err != nil {
		return "", err
	}
	tree, err := treeFromBlocks(blocks)
	if err != nil {
		return "", err
	}
	return renderPageDoc(tree)
}

// CreateExplicitPageVersionUseCase は「版を残す」操作（10 分規則を無視して必ず 1 件切る）。
// 本文そのものは変えないので TouchPageLastEditedBy は呼ばない。
type CreateExplicitPageVersionUseCase struct {
	versionRepo repository.PageVersionRepository
	kbRepo      repository.KnowledgeBaseRepository
	txManager   repository.TxManager
}

func NewCreateExplicitPageVersionUseCase(
	versionRepo repository.PageVersionRepository, kbRepo repository.KnowledgeBaseRepository, txManager repository.TxManager,
) *CreateExplicitPageVersionUseCase {
	return &CreateExplicitPageVersionUseCase{versionRepo: versionRepo, kbRepo: kbRepo, txManager: txManager}
}

type CreateExplicitPageVersionInput struct {
	WorkspaceID  string
	PageID       string
	AuthorUserID uint64
	// Note は版に添える任意のメモ。domain.ValidateVersionNote で検証・正規化してから保存する。
	Note *string
}

func (u *CreateExplicitPageVersionUseCase) Execute(ctx context.Context, in CreateExplicitPageVersionInput) (*domain.PageVersion, error) {
	// メモの検証は先に行う。不正なメモのためだけにトランザクションを開く無駄を避ける。
	note, err := domain.ValidateVersionNote(in.Note)
	if err != nil {
		return nil, err
	}
	var version *domain.PageVersion
	// pages 行のロック → 今の内容を読む → 版を挿入、をこの順で 1 トランザクションに入れる。
	// ロックより先に読むと、読み取りとロック取得の間に本物の編集が割り込み、古い内容のまま
	// 版を切ってしまう競合が実際に起きたことがある。
	err = u.txManager.DoInTx(ctx, func(ctx context.Context) error {
		if err := u.versionRepo.LockPage(ctx, in.WorkspaceID, in.PageID); err != nil {
			return err
		}
		doc, err := currentPageDoc(ctx, u.kbRepo, in.WorkspaceID, in.PageID)
		if err != nil {
			return err
		}
		// force=true — 10 分規則を無視して必ず切る（「版を残す」の定義そのもの）。
		// CreateVersionIfDue も同じ行を再ロックするが、同一トランザクション内の FOR UPDATE は
		// 再入可能なので待ちにはならない。
		_, v, err := u.versionRepo.CreateVersionIfDue(ctx, in.WorkspaceID, in.PageID, doc, in.AuthorUserID, note, true)
		if err != nil {
			return err
		}
		version = v
		return nil
	})
	if err != nil {
		return nil, err
	}
	return version, nil
}

// ListPageVersionsUseCase はページの版一覧（seq 降順）を返す。
type ListPageVersionsUseCase struct {
	versionRepo repository.PageVersionRepository
}

func NewListPageVersionsUseCase(versionRepo repository.PageVersionRepository) *ListPageVersionsUseCase {
	return &ListPageVersionsUseCase{versionRepo: versionRepo}
}

type ListPageVersionsInput struct {
	WorkspaceID string
	PageID      string
}

func (u *ListPageVersionsUseCase) Execute(ctx context.Context, in ListPageVersionsInput) ([]domain.PageVersion, error) {
	return u.versionRepo.ListVersions(ctx, in.WorkspaceID, in.PageID)
}

// GetPageVersionUseCase は版 1 件（doc 込み）を返す。
type GetPageVersionUseCase struct {
	versionRepo repository.PageVersionRepository
}

func NewGetPageVersionUseCase(versionRepo repository.PageVersionRepository) *GetPageVersionUseCase {
	return &GetPageVersionUseCase{versionRepo: versionRepo}
}

type GetPageVersionInput struct {
	WorkspaceID string
	PageID      string
	Seq         int64
}

func (u *GetPageVersionUseCase) Execute(ctx context.Context, in GetPageVersionInput) (*domain.PageVersion, error) {
	return u.versionRepo.GetVersion(ctx, in.WorkspaceID, in.PageID, in.Seq)
}

// RestorePageVersionUseCase は過去の版の doc を今の本文として書き戻す。
// ReplacePageBlocksUseCase をそのまま呼ぶ（本文保存と同じ検証・snapshot 焼き直し・最終編集者の
// 記録を経由する）。「復元自体も版になる」ため ForceVersion は必ず true。
type RestorePageVersionUseCase struct {
	versionRepo repository.PageVersionRepository
	replace     *ReplacePageBlocksUseCase
}

func NewRestorePageVersionUseCase(versionRepo repository.PageVersionRepository, replace *ReplacePageBlocksUseCase) *RestorePageVersionUseCase {
	return &RestorePageVersionUseCase{versionRepo: versionRepo, replace: replace}
}

type RestorePageVersionInput struct {
	WorkspaceID  string
	PageID       string
	Seq          int64
	EditorUserID uint64
}

func (u *RestorePageVersionUseCase) Execute(ctx context.Context, in RestorePageVersionInput) (*domain.PageSnapshot, error) {
	v, err := u.versionRepo.GetVersion(ctx, in.WorkspaceID, in.PageID, in.Seq)
	if err != nil {
		return nil, err
	}
	note := fmt.Sprintf("復元: 版%dから", in.Seq)
	return u.replace.Execute(ctx, ReplacePageBlocksInput{
		WorkspaceID:  in.WorkspaceID,
		PageID:       in.PageID,
		Doc:          v.Doc,
		EditorUserID: in.EditorUserID,
		ForceVersion: true,
		VersionNote:  &note,
	})
}
