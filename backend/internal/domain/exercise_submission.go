package domain

import "time"

// ExerciseSubmission は trainee がコード演習に提出したコード + 実行結果の 1 件（append-only）。
// 参照先テーブルは ExerciseKind で判定する polymorphic 設計（FK は張らずアプリ層で担保）。
type ExerciseSubmission struct {
	ID            uint64    `json:"id"`
	UserID        uint64    `json:"userId"`
	ExerciseKind  string    `json:"exerciseKind"`
	ExerciseID    uint64    `json:"exerciseId"`
	SubmittedCode string    `json:"submittedCode"`
	Stdout        string    `json:"stdout"`
	Stderr        string    `json:"stderr"`
	ExitCode      int       `json:"exitCode"`
	IsCorrect     bool      `json:"isCorrect"`
	SubmittedAt   time.Time `json:"submittedAt"`
}

// ExerciseKind* は ExerciseSubmission.ExerciseKind の許容値。
const (
	ExerciseKindMaster  = "master"
	ExerciseKindCompany = "company"
)
