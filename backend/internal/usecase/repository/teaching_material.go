package repository

import (
	"context"
	"errors"

	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// ErrChapterDocConflict は章のリッチ本文更新で revision が一致しなかった（他所で更新済み）。
var ErrChapterDocConflict = errors.New("chapter doc revision conflict")

// ErrChapterDocInvalidData は doc が PostgreSQL に格納できない不正データだった（U+0000 等）。
var ErrChapterDocInvalidData = errors.New("chapter doc contains invalid data")

// TeachingMaterialRepository は教材の永続化を担う（クエリは company_id 指定で他社漏れを防ぐ）。
type TeachingMaterialRepository interface {
	// ListByCompany は company 内全教材を返す backward-compat 用（コース対応への移行後に削除予定）。
	// 呼び出しは残っていないが、結合テストが直接使うため削除しない（ListByWorkspace が現行の経路）。
	ListByCompany(ctx context.Context, companyID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error)
	// ListByWorkspace はワークスペース内全教材を返す backward-compat 用（コース対応への移行後に削除予定）。
	// ListByCompany の workspace_id 版で、TeachingMaterialUseCase.List が使う現行の経路。
	ListByWorkspace(ctx context.Context, workspaceID string, includeUnpublished bool) ([]domain.TeachingMaterial, error)
	ListByCourse(ctx context.Context, courseID uint64, includeUnpublished bool) ([]domain.TeachingMaterial, error)
	GetByID(ctx context.Context, id uint64) (*domain.TeachingMaterial, error)
	// CountByCourseForWorkspace はワークスペース内の教材件数を course_id ごとに一括集計して返す
	// (コース一覧に章数を出すための集計。コースごとの個別クエリだと N+1 になるため)。
	// FRESTYLE-400 段4横展開: company_id 直読み（旧 CountByCourseForCompany）から
	// workspace_id 経由へ切り替え済み。
	CountByCourseForWorkspace(ctx context.Context, workspaceID string, includeUnpublished bool) (map[uint64]int, error)
	Create(ctx context.Context, m *domain.TeachingMaterial) error
	// Update は title / sort_order / is_published を書き換える。対象行が無ければ
	// domain.ErrNotFound（黙って成功にすると失われた編集を保存済みに見せるため）。
	Update(ctx context.Context, m *domain.TeachingMaterial) error
	// UpdateDocWithRevision はリッチ本文（tiptap JSON）を楽観ロックで更新する。
	// expectedRevision が現在値と一致した場合のみ doc を保存し revision を +1 する。
	// 不一致は ErrChapterDocConflict、未存在は domain.ErrNotFound を返す。
	UpdateDocWithRevision(ctx context.Context, id uint64, doc string, expectedRevision int) (*domain.TeachingMaterial, error)
	Delete(ctx context.Context, id uint64) error
	DeleteByCourse(ctx context.Context, courseID uint64) error
}
