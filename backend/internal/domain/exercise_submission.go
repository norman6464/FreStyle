package domain

import "time"

// ExerciseSubmission は trainee がコード演習に提出したコード + 実行結果の 1 件。
//
// 設計上のポイント:
//   - master_exercises / company_exercises は別テーブルなので polymorphic にする必要がある。
//     `ExerciseKind` で参照先テーブルを判定する（FK は張らずアプリ層で整合性を担保）。
//   - 提出は append-only で履歴として残す。trainee 1 人 × 1 問題に対し複数行を許容。
//   - ダッシュボード集計は (user_id, submitted_at DESC) インデックスで高速化。
//
// PR-H3 ではテーブル定義のみ。提出 API / 採点ロジックは別 PR で追加する。
type ExerciseSubmission struct {
	ID            uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        uint64    `gorm:"column:user_id;not null;index:idx_submissions_user_at,priority:1" json:"userId"`
	ExerciseKind  string    `gorm:"column:exercise_kind;size:16;not null" json:"exerciseKind"`
	ExerciseID    uint64    `gorm:"column:exercise_id;not null" json:"exerciseId"`
	SubmittedCode string    `gorm:"type:text;not null" json:"submittedCode"`
	Stdout        string    `gorm:"type:text" json:"stdout"`
	Stderr        string    `gorm:"type:text" json:"stderr"`
	ExitCode      int       `gorm:"not null;default:0" json:"exitCode"`
	IsCorrect     bool      `gorm:"not null;default:false" json:"isCorrect"`
	SubmittedAt   time.Time `gorm:"column:submitted_at;not null;index:idx_submissions_user_at,sort:desc,priority:2" json:"submittedAt"`
}

func (ExerciseSubmission) TableName() string { return "exercise_submissions" }

// ExerciseKind* は ExerciseSubmission.ExerciseKind の許容値。
const (
	ExerciseKindMaster  = "master"  // 運営共通の master_exercises を指す
	ExerciseKindCompany = "company" // 会社作成の company_exercises を指す
)
