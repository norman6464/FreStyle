package usecase

import (
	"fmt"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/repository"
)

// GetMasterExerciseDetailOutput は詳細ページに渡す問題本体 + 入出力例セット。
//
// JSON 表現は handler でそのままシリアライズされる。フロントは `examples` 配列を
// 入力例 1 / 入力例 2 / ... として全件描画し、提出時には同じ全件をテストケースとして
// 順に実行して合否を判定する（採点ロジックは PR-W で追加）。
type GetMasterExerciseDetailOutput struct {
	Exercise *domain.MasterExercise         `json:"exercise"`
	Examples []domain.MasterExerciseExample `json:"examples"`
}

// GetMasterExerciseUseCase は指定 slug の運営マスタ演習問題 + 入出力例を返す。
//
// 旧 API は `:id` ベースだったが、 paiza 風 URL `/code-editor/php-1` を実現するため
// 主導線を slug に切替える。 ID ベースの挙動も互換用に Execute / GetByID で残す。
type GetMasterExerciseUseCase struct {
	repo     repository.MasterExerciseRepository
	examples repository.MasterExerciseExampleRepository
}

func NewGetMasterExerciseUseCase(
	repo repository.MasterExerciseRepository,
	examples repository.MasterExerciseExampleRepository,
) *GetMasterExerciseUseCase {
	return &GetMasterExerciseUseCase{repo: repo, examples: examples}
}

// Execute は ID 指定で取得する旧 API 互換。 examples は付かない。
func (uc *GetMasterExerciseUseCase) Execute(id uint64) (*domain.MasterExercise, error) {
	return uc.repo.GetByID(id)
}

// ExecuteBySlug は paiza 風詳細ページ向け。 examples を含めて 1 度に返す。
//
// `GetBySlug` が GORM の `gorm.ErrRecordNotFound` を返した場合は handler 側で 404 に分岐できるよう
// そのまま伝搬する。 nil チェックは GORM が `(nil, nil)` を返さない前提だが、 リポジトリ実装の
// バグや fake で `ex == nil` が起きたときの nil pointer panic を防ぐため defensive に弾いておく。
func (uc *GetMasterExerciseUseCase) ExecuteBySlug(slug string) (*GetMasterExerciseDetailOutput, error) {
	ex, err := uc.repo.GetBySlug(slug)
	if err != nil {
		return nil, err
	}
	if ex == nil {
		return nil, fmt.Errorf("exercise not found: %s", slug)
	}
	examples, err := uc.examples.ListByExerciseID(ex.ID)
	if err != nil {
		return nil, err
	}
	return &GetMasterExerciseDetailOutput{Exercise: ex, Examples: examples}, nil
}
