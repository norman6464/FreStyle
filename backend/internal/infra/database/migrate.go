package database

import (
	"log"
	"os"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

// allDomainModels は全 domain 構造体のリスト。
// AutoMigrate に渡してテーブル作成・列追加を一元管理する。
// 新しい domain を追加したら必ずここにも追記する。
func allDomainModels() []any {
	return []any{
		&domain.User{},
		&domain.Profile{},
		&domain.AiChatSession{},
		&domain.PracticeScenario{},
		&domain.ScenarioBookmark{},
		&domain.SharedSession{},
		&domain.Note{},
		&domain.SessionNote{},
		&domain.ScoreCard{},
		&domain.ScoreGoal{},
		&domain.LearningReport{},
		&domain.ConversationTemplate{},
		&domain.FavoritePhrase{},
		&domain.Friendship{},
		&domain.Notification{},
		&domain.ReminderSetting{},
		&domain.DailyGoal{},
		&domain.WeeklyChallenge{},
		&domain.WeeklyChallengeProgress{},
		&domain.AdminInvitation{},
	}
}

// Migrate は起動時のスキーマ整合チェック。
//   - RESET_DB=true: DROP SCHEMA public CASCADE → CREATE SCHEMA public で完全 wipe してから AutoMigrate。
//     リリース前に Go domain を「正」とする初期構築をしたいときの一回限り操作。
//   - RESET_DB が未指定: AutoMigrate のみ走る（CREATE TABLE IF NOT EXISTS と ADD COLUMN は走るが、
//     型変更・列削除は GORM の安全側仕様で実行されない）。
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
	if err := db.AutoMigrate(allDomainModels()...); err != nil {
		return err
	}
	log.Println("migrate: AutoMigrate done")
	return nil
}
