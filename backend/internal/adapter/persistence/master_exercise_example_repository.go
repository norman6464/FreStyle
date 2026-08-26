package persistence

import (
	"context"
	"encoding/json"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence/sqlcgen"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/usecase/repository"
	"gorm.io/gorm"
)

// masterExerciseExampleRepository は [repository.MasterExerciseExampleRepository] の実装。
// クエリは sqlc 生成コード（生 SQL）で、GORM からは接続プール（*sql.DB）だけを借りる。
type masterExerciseExampleRepository struct {
	db *gorm.DB
}

func NewMasterExerciseExampleRepository(db *gorm.DB) repository.MasterExerciseExampleRepository {
	return &masterExerciseExampleRepository{db: db}
}

// toDomainExample は sqlc 生成モデル → domain への詰め替え。
// id 系は DB が bigint(int64) で持ち domain が uint64。値は採番シーケンス由来で常に非負・
// int64 範囲内のため変換は安全（gosec G115 は persistence の id 境界として .golangci.yml で除外）。
func toDomainExample(row sqlcgen.MasterExerciseExample) domain.MasterExerciseExample {
	return domain.MasterExerciseExample{
		ID:             uint64(row.ID),
		ExerciseID:     uint64(row.ExerciseID),
		OrderIndex:     row.OrderIndex,
		InputText:      row.InputText,
		ExpectedOutput: row.ExpectedOutput,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

func (r *masterExerciseExampleRepository) ListByExerciseID(ctx context.Context, exerciseID uint64) ([]domain.MasterExerciseExample, error) {
	// GORM の接続プールを共有して sqlc 生成クエリ（生 SQL）を実行する。
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).ListMasterExerciseExamplesByExerciseID(ctx, int64(exerciseID))
	if err != nil {
		return nil, err
	}
	examples := make([]domain.MasterExerciseExample, 0, len(rows))
	for _, row := range rows {
		examples = append(examples, toDomainExample(row))
	}
	return examples, nil
}

// ListByExerciseIDs は複数 exercise_id をまとめて取得し exercise_id ごとに map 化する（N+1 回避）。
func (r *masterExerciseExampleRepository) ListByExerciseIDs(ctx context.Context, exerciseIDs []uint64) (map[uint64][]domain.MasterExerciseExample, error) {
	result := make(map[uint64][]domain.MasterExerciseExample, len(exerciseIDs))
	if len(exerciseIDs) == 0 {
		return result, nil
	}
	ids := make([]int64, 0, len(exerciseIDs))
	for _, id := range exerciseIDs {
		if v, ok := toInt64ID(id); ok {
			ids = append(ids, v)
		}
	}
	if len(ids) == 0 {
		return result, nil // 存在し得ない id しか無い = 0 件
	}
	idsJSON, err := json.Marshal(ids)
	if err != nil {
		return nil, err
	}
	sqlDB, err := r.db.DB()
	if err != nil {
		return nil, err
	}
	rows, err := sqlcgen.New(sqlDB).ListMasterExerciseExamplesByExerciseIDs(ctx, idsJSON)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		ex := toDomainExample(row)
		result[ex.ExerciseID] = append(result[ex.ExerciseID], ex)
	}
	return result, nil
}
