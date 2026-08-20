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
		&domain.Role{},
		&domain.User{},
		&domain.UserOidcIdentity{},
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
	if err := SeedRoles(db); err != nil {
		return err
	}
	// users 正規化（FRESTYLE-311）のバックフィル。起動のたびに走るが冪等で、
	// 埋まっていれば no-op。デプロイと手動 SQL 適用の順序に依存させないためここで行う
	// （AutoMigrate → バックフィル → listen の順が 1 プロセス内で保証される）。
	if err := BackfillUserNormalization(db); err != nil {
		return err
	}
	if err := ApplyUserNormalizationConstraints(db); err != nil {
		return err
	}
	return nil
}

// SeedRoles はロールマスタを投入する（固定 ID・冪等）。起動時と結合テストのスキーマ構築で使う。
func SeedRoles(db *gorm.DB) error {
	seeds := []domain.Role{
		{ID: domain.RoleIDSuperAdmin, Name: domain.RoleSuperAdmin, Description: "運営管理者"},
		{ID: domain.RoleIDCompanyAdmin, Name: domain.RoleCompanyAdmin, Description: "企業管理者"},
		{ID: domain.RoleIDTrainee, Name: domain.RoleTrainee, Description: "受講者"},
	}
	for _, r := range seeds {
		if err := db.FirstOrCreate(&r, domain.Role{ID: r.ID}).Error; err != nil {
			return err
		}
	}
	return nil
}

// BackfillUserNormalization は旧カラム（users.role / users.cognito_sub）から
// 正規化テーブル（users.role_id / user_oidc_identities）へ値を埋める（冪等）。
func BackfillUserNormalization(db *gorm.DB) error {
	// role 文字列 → role_id。NULL 埋めに加え、移行期間中に旧コード（role 文字列のみ書く）が
	// 作った行の role_id 不一致も修復する（新コードは両方書くので通常は no-op）。
	if err := db.Exec(
		`UPDATE users SET role_id = r.id FROM roles r
		 WHERE r.name = users.role AND users.role_id IS DISTINCT FROM r.id`,
	).Error; err != nil {
		return err
	}
	// role が空・未知のままの行は最小権限の trainee に倒す（読み出し側の LEFT JOIN で NULL role を作らない）。
	if err := db.Exec(
		`UPDATE users SET role_id = ? WHERE role_id IS NULL`, domain.RoleIDTrainee,
	).Error; err != nil {
		return err
	}
	// cognito_sub → user_oidc_identities（既存行はスキップ）。
	return db.Exec(
		`INSERT INTO user_oidc_identities (user_id, provider, subject, created_at, updated_at)
		 SELECT id, ?, cognito_sub, NOW(), NOW() FROM users
		 WHERE cognito_sub IS NOT NULL AND cognito_sub <> ''
		 ON CONFLICT DO NOTHING`, domain.OidcProviderCognito,
	).Error
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

// ApplyUserNormalizationConstraints は正規化テーブルの整合性制約を適用する（冪等・FRESTYLE-311）。
// バックフィル後に呼ぶ前提（既存行が制約を満たしてから付ける）。GORM の AutoMigrate は
// FK / CHECK / 部分 UNIQUE を表現できないため、ここで明示 SQL として管理する。
func ApplyUserNormalizationConstraints(db *gorm.DB) error {
	stmts := []string{
		// roles.name: 空文字禁止。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_roles_name_not_empty') THEN
				ALTER TABLE roles ADD CONSTRAINT ck_roles_name_not_empty CHECK (name <> '');
			END IF;
		END $$;`,
		// users.role_id: 既定値 trainee(3)。移行期間中の旧コード（role_id を書かない INSERT）を
		// NOT NULL 違反で壊さないための安全弁。既定で入った値は次回起動のバックフィル
		// （role 文字列との不一致修復）が正しいロールへ補正する。
		`ALTER TABLE users ALTER COLUMN role_id SET DEFAULT 3;`,
		// users.role_id: バックフィル済み + DEFAULT ありなので NOT NULL にできる。
		`DO $$ BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'users' AND column_name = 'role_id' AND is_nullable = 'YES'
			) THEN
				ALTER TABLE users ALTER COLUMN role_id SET NOT NULL;
			END IF;
		END $$;`,
		// users.role_id → roles.id。ロールマスタの行は参照されている限り消せない（RESTRICT 相当）。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_users_role') THEN
				ALTER TABLE users ADD CONSTRAINT fk_users_role FOREIGN KEY (role_id) REFERENCES roles(id);
			END IF;
		END $$;`,
		// user_oidc_identities.user_id → users.id。ユーザーの物理削除で identity も消す。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_user_oidc_identities_user') THEN
				ALTER TABLE user_oidc_identities
					ADD CONSTRAINT fk_user_oidc_identities_user
					FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
			END IF;
		END $$;`,
		// identity の provider / subject: 空文字禁止。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_user_oidc_identities_not_empty') THEN
				ALTER TABLE user_oidc_identities
					ADD CONSTRAINT ck_user_oidc_identities_not_empty CHECK (provider <> '' AND subject <> '');
			END IF;
		END $$;`,
		// users.email: アプリは必ず値を入れる（Go の string 非ポインタ）ため NOT NULL にする。
		// 万一 NULL 行が残っている場合は WARNING に留め、起動は落とさない。
		`DO $$ BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'users' AND column_name = 'email' AND is_nullable = 'YES'
			) THEN
				IF EXISTS (SELECT 1 FROM users WHERE email IS NULL) THEN
					RAISE WARNING 'users.email に NULL があるため NOT NULL を適用できません';
				ELSE
					ALTER TABLE users ALTER COLUMN email SET NOT NULL;
				END IF;
			END IF;
		END $$;`,
		// users.email: アクティブ行（未論理削除）に限った部分 UNIQUE。論理削除→同メール再招待と両立する。
		// 既存データに重複がある場合は作成せず WARNING を出す（起動を落とさず、修正は運用判断に委ねる）。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'uq_users_email_active') THEN
				IF EXISTS (
					SELECT email FROM users WHERE deleted_at IS NULL AND email <> ''
					GROUP BY email HAVING count(*) > 1
				) THEN
					RAISE WARNING 'users.email に重複があるため uq_users_email_active を作成できません（重複を解消して再起動してください）';
				ELSE
					CREATE UNIQUE INDEX uq_users_email_active ON users (email) WHERE deleted_at IS NULL;
				END IF;
			END IF;
		END $$;`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}
	return nil
}
