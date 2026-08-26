package domain

import "time"

// MasterExerciseExample は MasterExercise に紐付く入力例 / 期待出力例の 1 ペア。
// 1 問に複数ケースを持ち、表示 / 採点順序は OrderIndex で安定ソートする。
type MasterExerciseExample struct {
	ID uint64 `json:"id"`
	// (exercise_id, order_index) の UNIQUE で同一問題内の OrderIndex 衝突を DB レベルで弾く。
	ExerciseID uint64 `json:"exerciseId"`
	// DEFAULT を持たせない（default:0 だと未指定 INSERT が 0 で衝突するため）。
	OrderIndex     int16     `json:"orderIndex"`
	InputText      string    `json:"inputText"`
	ExpectedOutput string    `json:"expectedOutput"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}
