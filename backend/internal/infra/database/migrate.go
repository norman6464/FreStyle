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
		// リッチテキスト文書（tiptap JSON を jsonb で保持）。
		&domain.RichDocument{},
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
	// AutoMigrate が email / role_id へ NOT NULL を張るため、旧スキーマの NULL を先に埋める
	// （NULL が残っていると SET NOT NULL で起動が落ちるため）。冪等。
	if err := preRepairUsersForMigrate(db); err != nil {
		return err
	}
	log.Println("migrate: AutoMigrate start")
	if err := AutoMigrateAll(db); err != nil {
		return err
	}
	log.Println("migrate: AutoMigrate done")
	// 演習データ(PHP / Go / Docker / Linux / Git など)は問題文・期待出力を公開リポに露出させない
	// ため本体には埋め込まず、非公開の教材リポ(frestyle-teaching-materials/exercises/<lang>/*.md)を
	// 唯一の正本とし、seed.py が生成する UPSERT SQL を Supabase に流して投入する。
	// seed / バックフィル / 制約適用は check-then-act を含むため、複数タスクの同時起動でも
	// 直列化されるよう advisory lock で囲む。pgbouncer(transaction pooler) 前提のため
	// セッションロックではなくトランザクションロック（コミットで自動解放）を使う。
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(`SELECT pg_advisory_xact_lock(4915311)`).Error; err != nil {
			return err
		}
		if err := seedCompanies(tx); err != nil {
			return err
		}
		if err := SeedRoles(tx); err != nil {
			return err
		}
		// users 正規化（FRESTYLE-311）のバックフィル。起動のたびに走るが冪等で、
		// 埋まっていれば no-op。デプロイと手動 SQL 適用の順序に依存させないためここで行う
		// （AutoMigrate → バックフィル → listen の順が 1 プロセス内で保証される）。
		if err := BackfillUserNormalization(tx); err != nil {
			return err
		}
		if err := ApplyUserNormalizationConstraints(tx); err != nil {
			return err
		}
		return ApplyRichDocumentConstraints(tx)
	})
}

// ApplyRichDocumentConstraints は rich_documents の整合性制約を適用する（冪等）。
// GORM の AutoMigrate は FK / CHECK を表現できないため、ここで明示 SQL として管理する。
func ApplyRichDocumentConstraints(db *gorm.DB) error {
	stmts := []string{
		// owner_id → users.id。ユーザーの物理削除で文書も消す（論理削除運用なので通常は発火しない）。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_rich_documents_owner') THEN
				ALTER TABLE rich_documents
					ADD CONSTRAINT fk_rich_documents_owner
					FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE;
			END IF;
		END $$;`,
		// doc は tiptap のドキュメント JSON（object かつ type='doc'）に限る。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_rich_documents_doc') THEN
				ALTER TABLE rich_documents
					ADD CONSTRAINT ck_rich_documents_doc
					CHECK (jsonb_typeof(doc) = 'object' AND doc->>'type' = 'doc');
			END IF;
		END $$;`,
		// title 長の上限（アプリ側検証と二重の壁）。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_rich_documents_title_len') THEN
				ALTER TABLE rich_documents
					ADD CONSTRAINT ck_rich_documents_title_len
					CHECK (char_length(title) <= 200);
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

// preRepairUsersForMigrate は AutoMigrate（NOT NULL 適用）の前提を満たすよう旧データを埋める（冪等）。
// users テーブル / 対象カラムが未作成の初回起動では no-op。
func preRepairUsersForMigrate(db *gorm.DB) error {
	return db.Exec(`DO $$ BEGIN
		IF to_regclass('users') IS NOT NULL THEN
			UPDATE users SET email = '' WHERE email IS NULL;
			IF EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'users' AND column_name = 'role_id'
			) THEN
				UPDATE users SET role_id = 3 WHERE role_id IS NULL;
			END IF;
		END IF;
	END $$;`).Error
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

// BackfillUserNormalization は正規化テーブル（users.role_id / user_oidc_identities）の
// 整合を起動時に保つ（冪等）。正規化後は role_id が正で新コードが常に書くため、旧 role 文字列から
// role_id を「逆算」する同期は行わない（それをやると、role_id だけ更新した昇格を巻き戻す）。
// 旧カラム users.cognito_sub からの identity 補完だけは、旧カラム撤去（migrations/0021）の前後で
// 安全に流せるよう、カラムが存在する間のみカラム存在チェックでガードして実行する。
func BackfillUserNormalization(db *gorm.DB) error {
	// role_id が未設定の行は最小権限の trainee に倒す（読み出し側の LEFT JOIN で NULL role を作らない）。
	// role_id のみを触るため旧カラム撤去後も有効。
	if err := db.Exec(
		`UPDATE users SET role_id = ? WHERE role_id IS NULL`, domain.RoleIDTrainee,
	).Error; err != nil {
		return err
	}
	// cognito_sub → user_oidc_identities（旧カラムが残っている間のみ・既存行はスキップ）。
	// 論理削除済みユーザーは対象外（identity が subject を占有すると再招待がログイン不能になる）。
	hasCognitoSub, err := columnExists(db, "users", "cognito_sub")
	if err != nil {
		return err
	}
	if hasCognitoSub {
		if err := db.Exec(
			`INSERT INTO user_oidc_identities (user_id, provider, subject, created_at, updated_at)
			 SELECT id, ?, cognito_sub, NOW(), NOW() FROM users
			 WHERE cognito_sub IS NOT NULL AND cognito_sub <> '' AND deleted_at IS NULL
			 ON CONFLICT DO NOTHING`, domain.OidcProviderCognito,
		).Error; err != nil {
			return err
		}
	}
	// 論理削除済みユーザーに紐付く identity を掃除する（SoftDelete 側でも消すが、
	// 過去データと削除処理の失敗に対する自己修復として毎起動流す。冪等）。
	return db.Exec(
		`DELETE FROM user_oidc_identities oi USING users u
		 WHERE oi.user_id = u.id AND u.deleted_at IS NOT NULL`,
	).Error
}

// columnExists は指定テーブルにカラムが存在するかを返す。旧カラム撤去の前後で
// バックフィルの分岐に使う（information_schema はトランザクション内でも現在のスキーマを見る）。
func columnExists(db *gorm.DB, table, column string) (bool, error) {
	var n int64
	if err := db.Raw(
		`SELECT count(*) FROM information_schema.columns
		 WHERE table_name = ? AND column_name = ?`, table, column,
	).Scan(&n).Error; err != nil {
		return false, err
	}
	return n > 0, nil
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
		// users.role_id の NOT NULL / DEFAULT は domain タグ経由で AutoMigrate が管理する
		// （ここで別途 ALTER すると AutoMigrate が毎起動剥がして貼り直す羽目になる）。
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
		// users.email の NOT NULL も domain タグ経由で AutoMigrate が管理する
		// （preRepairUsersForMigrate が NULL を先に埋めるため適用は常に成功する）。
		// users.email: アクティブ行（未論理削除）かつ非空に限った部分 UNIQUE。
		// 論理削除→同メール再招待と両立し、email claim の無い OIDC ユーザー（空文字）は対象外にする
		// （重複ガードと述語を必ず一致させること。ずれると起動失敗が自己修復しなくなる）。
		// 既存データに重複がある場合は作成せず WARNING を出す（起動を落とさず、修正は運用判断に委ねる）。
		`DO $$ BEGIN
			IF EXISTS (
				SELECT 1 FROM pg_indexes WHERE indexname = 'uq_users_email_active'
				  AND indexdef NOT LIKE '%email%<>%'
			) THEN
				DROP INDEX uq_users_email_active;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'uq_users_email_active') THEN
				IF EXISTS (
					SELECT email FROM users WHERE deleted_at IS NULL AND email <> ''
					GROUP BY email HAVING count(*) > 1
				) THEN
					RAISE WARNING 'users.email に重複があるため uq_users_email_active を作成できません（重複を解消して再起動してください）';
				ELSE
					CREATE UNIQUE INDEX uq_users_email_active
						ON users (email) WHERE deleted_at IS NULL AND email <> '';
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
