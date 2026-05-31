package domain

import "time"

// MasterExerciseExample は MasterExercise に紐付く入力例 / 期待出力例の 1 ペア。
// 1 問に複数ケースを持ち、表示 / 採点順序は OrderIndex で安定ソートする。
type MasterExerciseExample struct {
	ID uint64 `gorm:"primaryKey;autoIncrement" json:"id"`
	// (exercise_id, order_index) の UNIQUE で同一問題内の OrderIndex 衝突を DB レベルで弾く。
	ExerciseID uint64 `gorm:"column:exercise_id;not null;uniqueIndex:idx_examples_exercise_order,priority:1" json:"exerciseId"`
	// DEFAULT を持たせない（default:0 だと未指定 INSERT が 0 で衝突するため）。
	OrderIndex     int16     `gorm:"column:order_index;type:smallint;not null;uniqueIndex:idx_examples_exercise_order,priority:2" json:"orderIndex"`
	InputText      string    `gorm:"column:input_text;type:text;not null;default:''" json:"inputText"`
	ExpectedOutput string    `gorm:"column:expected_output;type:text;not null" json:"expectedOutput"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (MasterExerciseExample) TableName() string { return "master_exercise_examples" }
