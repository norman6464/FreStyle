//go:build integration

package database_test

import (
	"testing"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"github.com/norman6464/FreStyle/backend/internal/infra/database"
	"github.com/norman6464/FreStyle/backend/internal/testsupport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSeedCleanArchitectureExercise_Integration は Migrate がクリーンアーキテクチャの
// Go 演習を投入し、冪等（再実行しても 1 件）であることを実 Postgres で検証する。
func TestSeedCleanArchitectureExercise_Integration(t *testing.T) {
	db := testsupport.OpenTestDB(t)
	require.NoError(t, database.Migrate(db))

	const slug = "go-clean-arch-greeting"
	var ex domain.MasterExercise
	require.NoError(t, db.Where("slug = ?", slug).First(&ex).Error)

	assert.Equal(t, domain.ExerciseLanguageGo, ex.Language)
	assert.Equal(t, "アーキテクチャ", ex.Category)
	assert.Equal(t, domain.ExerciseModeExecute, ex.Mode)
	assert.Equal(t, "Hello, FreStyle! (clean architecture)", ex.ExpectedOutput)
	// クリーンアーキの構造（port / usecase）が starter に含まれる。
	assert.Contains(t, ex.StarterCode, "GreetingRepository")
	assert.Contains(t, ex.StarterCode, "GreetUseCase")

	// example（採点用テストケース）が 1 件できている。
	var exampleCount int64
	require.NoError(t, db.Model(&domain.MasterExerciseExample{}).
		Where("exercise_id = ?", ex.ID).Count(&exampleCount).Error)
	assert.Equal(t, int64(1), exampleCount)

	// 冪等: 再度 Migrate しても clean-arch 演習は 1 件のまま。
	require.NoError(t, database.Migrate(db))
	var caCount int64
	require.NoError(t, db.Model(&domain.MasterExercise{}).
		Where("slug = ?", slug).Count(&caCount).Error)
	assert.Equal(t, int64(1), caCount)
}
