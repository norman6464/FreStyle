package database

import (
	"log"
	"os"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

// allDomainModels は AutoMigrate に渡す全 domain 構造体のリスト。
// 新しい domain を追加したらここにも追記する。
func allDomainModels() []any {
	return []any{
		&domain.User{},
		&domain.Profile{},
		&domain.AiChatSession{},
		&domain.Note{},
		&domain.SessionNote{},
		&domain.Notification{},
		&domain.AdminInvitation{},
		&domain.MasterExercise{},
		&domain.MasterExerciseExample{},
		&domain.CompanyExercise{},
		&domain.ExerciseSubmission{},
		&domain.Company{},
		&domain.CompanyApplication{},
		&domain.Course{},
		&domain.TeachingMaterial{},
		&domain.LearningReport{},
		&domain.AuditEvent{},
		// UserLessonProgress のテーブルは user_chapter_progress(FRESTYLE-186 で移行完了)。
		&domain.UserLessonProgress{},
		// user_chapter_views / user_daily_activities の実テーブルは migration 0005 で作成済。
		// ここに載せるのは結合テスト DB のスキーマ構築のため(タグは 0005 と一致させ、本番では no-op)。
		&domain.UserChapterView{},
		&domain.UserDailyActivity{},
	}
}

// AutoMigrateAll は全 domain モデルを AutoMigrate する（seed なし）。
// 起動時の Migrate と、結合テストのスキーマ初期化の両方から使う（モデル一覧の単一情報源）。
func AutoMigrateAll(db *gorm.DB) error {
	return db.AutoMigrate(allDomainModels()...)
}

// Migrate は起動時にスキーマを AutoMigrate する。
// RESET_DB=true のときは public schema を完全 wipe してから再構築する（一回限りの初期構築用）。
func Migrate(db *gorm.DB) error {
	if os.Getenv("RESET_DB") == "true" {
		log.Println("⚠️ RESET_DB=true: dropping public schema and recreating")
		if err := db.Exec("DROP SCHEMA public CASCADE").Error; err != nil {
			return err
		}
		if err := db.Exec("CREATE SCHEMA public").Error; err != nil {
			return err
		}
	}
	log.Println("migrate: AutoMigrate start")
	if err := AutoMigrateAll(db); err != nil {
		return err
	}
	log.Println("migrate: AutoMigrate done")
	// 演習データ(PHP / Go / Docker / Linux / Git など)は問題文・期待出力を公開リポに露出させない
	// ため本体には埋め込まず、非公開の教材リポ(frestyle-teaching-materials/exercises/<lang>/*.md)を
	// 唯一の正本とし、seed.py が生成する UPSERT SQL を Supabase に流して投入する。
	if err := seedCompanies(db); err != nil {
		return err
	}
	return nil
}

func seedCompanies(db *gorm.DB) error {
	seeds := []domain.Company{
		{ID: 1, Name: "株式会社FreStyle"},
	}
	for _, c := range seeds {
		if err := db.FirstOrCreate(&c, domain.Company{ID: c.ID}).Error; err != nil {
			return err
		}
	}
	return nil
}
