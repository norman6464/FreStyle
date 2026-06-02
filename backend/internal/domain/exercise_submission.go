package domain

import "time"

// ExerciseSubmission は trainee がコード演習に提出したコード + 実行結果の 1 件（append-only）。
// 参照先テーブルは ExerciseKind で判定する polymorphic 設計（FK は張らずアプリ層で担保）。
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
	ExerciseKindMaster  = "master"
	ExerciseKindCompany = "company"
)
