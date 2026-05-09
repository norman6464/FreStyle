package domain

import "time"

// MasterExerciseExample は MasterExercise に紐付く「入力例 / 期待出力例」の 1 ペア。
//
// 1 つの問題に対して複数のテストケース（入力例 1, 入力例 2, ...）を表現する。
// 採点はこの example 全件をコードに食わせて期待出力と一致するか比較するロジックで行う
// （採点 usecase は別 PR で実装予定）。
//
// 設計上のポイント:
//   - ExerciseID は master_exercises への FK だが、 GORM Constraint でカスケード削除を強制する
//     よりはアプリ層 / マイグレーションで明示削除する方針（運営は基本的に問題を物理削除しない）
//   - InputText は標準入力に流す内容。問題によっては入力なし（空文字）でも成立する
//   - 表示 / 採点順序は OrderIndex で安定ソート
type MasterExerciseExample struct {
	ID             uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ExerciseID     uint64    `gorm:"column:exercise_id;not null;index:idx_examples_exercise_order,priority:1" json:"exerciseId"`
	OrderIndex     int16     `gorm:"column:order_index;type:smallint;not null;default:0;index:idx_examples_exercise_order,priority:2" json:"orderIndex"`
	InputText      string    `gorm:"column:input_text;type:text;not null;default:''" json:"inputText"`
	ExpectedOutput string    `gorm:"column:expected_output;type:text;not null" json:"expectedOutput"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (MasterExerciseExample) TableName() string { return "master_exercise_examples" }
