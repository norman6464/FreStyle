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
	sqlDB := testsupport.SQLDB(t, db)
	repo := persistence.NewMasterExerciseExampleRepository(sqlDB)
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

// TestMasterExerciseExampleRepository_ListByExerciseIDs_Integration は、複数 exercise_id を
// まとめて取得する ListByExerciseIDs（N+1 回避）を実 Postgres で検証する。
// exercise_id ごとに map 化され、各スライスは (order_index, id) 昇順に並ぶことを固定する。
func TestMasterExerciseExampleRepository_ListByExerciseIDs_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	sqlDB := testsupport.SQLDB(t, db)
	repo := persistence.NewMasterExerciseExampleRepository(sqlDB)
	ctx := context.Background()
	testsupport.TruncateAll(t, db, "master_exercise_examples")

	seed := []domain.MasterExerciseExample{
		{ExerciseID: 10, OrderIndex: 2, InputText: "b", ExpectedOutput: "B"},
		{ExerciseID: 10, OrderIndex: 1, InputText: "a", ExpectedOutput: "A"},
		{ExerciseID: 20, OrderIndex: 1, InputText: "x", ExpectedOutput: "X"},
		{ExerciseID: 99, OrderIndex: 1, InputText: "z", ExpectedOutput: "Z"}, // 要求しない問題
	}
	for i := range seed {
		require.NoError(t, db.WithContext(ctx).Create(&seed[i]).Error)
	}

	got, err := repo.ListByExerciseIDs(ctx, []uint64{10, 20})
	require.NoError(t, err)
	require.Len(t, got, 2, "要求した 2 問だけが map に入る（99 は含まれない）")

	require.Len(t, got[10], 2)
	require.Equal(t, int16(1), got[10][0].OrderIndex, "order_index 昇順")
	require.Equal(t, "a", got[10][0].InputText)
	require.Equal(t, int16(2), got[10][1].OrderIndex)
	require.Equal(t, uint64(10), got[10][0].ExerciseID)

	require.Len(t, got[20], 1)
	require.Equal(t, "x", got[20][0].InputText)

	_, ok := got[99]
	require.False(t, ok, "要求していない exercise_id は map に現れない")

	t.Run("空スライスは空 map を返しクエリを打たない", func(t *testing.T) {
		empty, err := repo.ListByExerciseIDs(ctx, nil)
		require.NoError(t, err)
		require.Empty(t, empty)
	})
}
