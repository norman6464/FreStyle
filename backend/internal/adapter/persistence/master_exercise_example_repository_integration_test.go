//go:build integration

package persistence_test

import (
	"context"
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/adapter/persistence"
	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/require"
)

// TestMasterExerciseExampleRepository_ListByExerciseID_Integration は、sqlc 生成クエリ（生 SQL）に
// 置き換えた ListByExerciseID を実 Postgres で検証する。GORM の *sql.DB を共有して動くことの担保。
func TestMasterExerciseExampleRepository_ListByExerciseID_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	repo := persistence.NewMasterExerciseExampleRepository(db)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "master_exercise_examples")

	// order_index を意図的に降順で投入し、クエリ側で order_index 昇順に並ぶことを確認する。
	seed := []domain.MasterExerciseExample{
		{ExerciseID: 10, OrderIndex: 2, InputText: "b", ExpectedOutput: "B"},
		{ExerciseID: 10, OrderIndex: 1, InputText: "a", ExpectedOutput: "A"},
		{ExerciseID: 99, OrderIndex: 1, InputText: "x", ExpectedOutput: "X"},
	}
	for i := range seed {
		require.NoError(t, db.WithContext(ctx).Create(&seed[i]).Error)
	}

	rows, err := repo.ListByExerciseID(ctx, 10)
	require.NoError(t, err)
	require.Len(t, rows, 2, "exercise_id=10 の 2 件のみ（99 は別問題）")
	require.Equal(t, int16(1), rows[0].OrderIndex, "order_index 昇順")
	require.Equal(t, "a", rows[0].InputText)
	require.Equal(t, int16(2), rows[1].OrderIndex)
	require.Equal(t, uint64(10), rows[0].ExerciseID)
}
