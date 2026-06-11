package database

import (
	"log"
	"time"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

// seedMasterExercises は埋め込みの初期演習データを投入する。
//
// PHP / Go / Docker / Linux / Git などの言語別演習は、問題文・期待出力を公開リポに露出させない
// ため本体（公開リポ）には埋め込まず、非公開の教材リポ
// (frestyle-teaching-materials/exercises/<lang>/*.md) を唯一の正本とする。
// 教材リポの seed.py が slug をキーに UPSERT する SQL を生成し、Supabase に流して投入する。
// ここでは本体コードと密結合な「クリーンアーキテクチャ体験用 Go 演習」のみ埋め込みで投入する。
func seedMasterExercises(db *gorm.DB) error {
	return seedCleanArchitectureExercise(db)
}

// seedCleanArchitectureExercise はクリーンアーキテクチャ（依存性逆転）を 1 ファイルの Go で
// 体験する演習を投入する。slug の存在チェックで冪等（autoIncrement で ID を採番し、その ID で
// example を 1 件作る）。
func seedCleanArchitectureExercise(db *gorm.DB) error {
	const slug = "go-clean-arch-greeting"

	var count int64
	if err := db.Model(&domain.MasterExercise{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now()
	const expected = "Hello, FreStyle! (clean architecture)"
	ex := domain.MasterExercise{
		Slug:       slug,
		Language:   domain.ExerciseLanguageGo,
		OrderIndex: 1000,
		Category:   "アーキテクチャ",
		Title:      "クリーンアーキテクチャ：依存性逆転で挨拶ユースケースを実装",
		Description: "1 つのファイルでクリーンアーキテクチャの肝である **依存性逆転 (DIP)** を体験します。\n\n" +
			"このコードは 4 つの役割に分かれています:\n\n" +
			"- **domain（エンティティ）**: `Greeting`（業務データ）\n" +
			"- **port（インターフェース）**: `GreetingRepository`。usecase が依存する抽象\n" +
			"- **usecase（アプリケーションロジック）**: `GreetUseCase`。具体実装ではなく port に依存する\n" +
			"- **infra（実装）**: `InMemoryGreetingRepository`。port を満たす\n" +
			"- **main（wiring）**: 依存を組み立てて usecase に注入する\n\n" +
			"`GreetUseCase.Execute` が未実装（`return \"\"`）です。**repo から `Greeting` を取得し、その `Message` を返す**ように実装してください。\n\n" +
			"期待する出力:\n\n```\n" + expected + "\n```",
		StarterCode: `package main

import "fmt"

// ===== domain（エンティティ）=====
type Greeting struct {
	Message string
}

// ===== port（usecase が依存するインターフェース）=====
// 依存性逆転(DIP): usecase は具体実装ではなく、この抽象に依存する。
type GreetingRepository interface {
	FindByName(name string) Greeting
}

// ===== usecase（アプリケーションロジック）=====
type GreetUseCase struct {
	repo GreetingRepository
}

func NewGreetUseCase(repo GreetingRepository) *GreetUseCase {
	return &GreetUseCase{repo: repo}
}

func (uc *GreetUseCase) Execute(name string) string {
	// TODO: repo から Greeting を取得し、その Message を返す
	return ""
}

// ===== infra（port の実装）=====
type InMemoryGreetingRepository struct{}

func (r *InMemoryGreetingRepository) FindByName(name string) Greeting {
	return Greeting{Message: "Hello, " + name + "! (clean architecture)"}
}

// ===== main（wiring: 依存を組み立てて注入）=====
func main() {
	repo := &InMemoryGreetingRepository{}
	uc := NewGreetUseCase(repo)
	fmt.Println(uc.Execute("FreStyle"))
}
`,
		HintText: "`Execute` の中で `uc.repo.FindByName(name)` を呼び、返ってきた `Greeting` の `.Message` を `return` します。" +
			"usecase は `InMemoryGreetingRepository` を直接知らず、`GreetingRepository`（port）越しに使う点が依存性逆転です。",
		ExpectedOutput: expected,
		Explanation: "usecase（`GreetUseCase`）は具体実装ではなく `GreetingRepository`（port）にだけ依存しています。" +
			"実装（`InMemoryGreetingRepository`）は `main` で注入されるため、後から DB 実装や別実装に差し替えても usecase は変更不要です。" +
			"これが「依存は内側（抽象）へ向ける」クリーンアーキテクチャの核心です。",
		Mode:        domain.ExerciseModeExecute,
		Difficulty:  3,
		IsPublished: true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(&ex).Error; err != nil {
		return err
	}

	example := domain.MasterExerciseExample{
		ExerciseID:     ex.ID,
		OrderIndex:     1,
		InputText:      "",
		ExpectedOutput: expected,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := db.Create(&example).Error; err != nil {
		return err
	}
	log.Printf("seed: clean-architecture Go 演習を挿入（slug=%s, id=%d）", slug, ex.ID)
	return nil
}
