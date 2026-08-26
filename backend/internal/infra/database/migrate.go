package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/norman6464/FreStyle/backend/internal/domain"
)

// migrateAdvisoryLockKey は起動時マイグレーションを直列化する advisory lock のキー。
// 複数の ECS タスクが同時に起動しても、DDL / seed / バックフィル / 制約適用が
// 同じキーの下で 1 つずつ流れるようにする。
const migrateAdvisoryLockKey = 4915311

const (
	// migrateLockTimeout は起動時マイグレーションがテーブル等のロック取得を待つ上限。
	// 未設定（PostgreSQL の既定は無制限）だと、長い書き込みトランザクションが 1 本あるだけで
	// マイグレーションが永久に待ち、その後ろに全ライターが積み上がる。
	migrateLockTimeout = "3s"

	// migrateAdvisoryLockTimeout は先行タスクの起動時マイグレーション完了を待つ上限。
	// 待つ相手がアプリのライターではなく自分たちの前段なので、テーブルロックより長く取る
	// （ローリングデプロイでは先行タスクの DDL が終わるまで待つのが正しい振る舞い）。
	migrateAdvisoryLockTimeout = "30s"

	// migrateLockRetries はロック待ちがタイムアウトしたときに同じ段をやり直す回数。
	migrateLockRetries = 4

	// lockNotAvailableCode は lock_timeout 超過を表す PostgreSQL の SQLSTATE（55P03）。
	lockNotAvailableCode = "55P03"
)

// Executor は *sql.DB と *sql.Tx が共通で満たす実行インターフェース。
// 起動時は 1 つのトランザクションへまとめて流し、結合テストは接続プールへ直接流すため、
// seed / バックフィル / 制約適用はどちらでも呼べるようにこの形で受ける。
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// Migrate は起動時にスキーマを適用し、seed / バックフィル / 制約適用まで済ませる。
// RESET_DB=true のときは public schema を完全 wipe してから再構築する（一回限りの初期構築用）。
//
// 適用順序は依存関係で決まっており崩せない:
//
//	中核スキーマ → seed / バックフィル / 明示制約 → ナレッジ基盤 → 権限モデル → テナント橋渡し
//
// 権限モデルは users を、テナント橋渡しは workspaces を参照する。
func Migrate(ctx context.Context, db *sql.DB) error {
	// スキーマの作り直しは他タスクの DDL 適用と重なると壊れるため、後段と同じ advisory lock で
	// 直列化する。ロックを取らずに DROP すると、先行タスクが CREATE 中のオブジェクトを
	// 消してしまう。
	if os.Getenv("RESET_DB") == "true" {
		if err := withMigrateTx(ctx, db, "スキーマの作り直し", func(tx *sql.Tx) error {
			log.Println("⚠️ RESET_DB=true: dropping public schema and recreating")
			if _, err := tx.ExecContext(ctx, "DROP SCHEMA public CASCADE"); err != nil {
				return err
			}
			_, err := tx.ExecContext(ctx, "CREATE SCHEMA public")
			return err
		}); err != nil {
			return err
		}
	}
	// 旧スキーマの NULL を先に埋める。正規化前の行が残っている環境でも、後段のバックフィルと
	// 部分 UNIQUE が素直に通るようにするための下ごしらえ。冪等。
	// lock_timeout と直列化を効かせるため、この UPDATE も withMigrateTx の中で流す
	// （素の autocommit だと、ロック待ちが無制限のまま users を掴みに行く）。
	if err := withMigrateTx(ctx, db, "旧データの下ごしらえ", func(tx *sql.Tx) error {
		return preRepairUsersForMigrate(ctx, tx)
	}); err != nil {
		return err
	}
	log.Println("migrate: core schema start")
	if err := ApplyCoreSchema(ctx, db); err != nil {
		return err
	}
	log.Println("migrate: core schema done")
	// 演習データ(PHP / Go / Docker / Linux / Git など)は問題文・期待出力を公開リポに露出させない
	// ため本体には埋め込まず、非公開の教材リポ(frestyle-teaching-materials/exercises/<lang>/*.md)を
	// 唯一の正本とし、seed.py が生成する UPSERT SQL を Supabase に流して投入する。
	// seed / バックフィル / 制約適用は check-then-act を含むため、複数タスクの同時起動でも
	// 直列化されるよう advisory lock で囲む。
	if err := withMigrateTx(ctx, db, "seed とバックフィル", func(tx *sql.Tx) error {
		if err := seedCompanies(ctx, tx); err != nil {
			return err
		}
		if err := SeedRoles(ctx, tx); err != nil {
			return err
		}
		// seed は固定 id で INSERT するため採番列（シーケンス）が進まない。そのままだと
		// 次の「id を書かない INSERT」が既存 id を引き当てて主キー衝突で落ちる。
		// seed の直後に実際の最大 id へ合わせておく（詳細は syncSeededSequences のコメント）。
		if err := syncSeededSequences(ctx, tx); err != nil {
			return err
		}
		// users 正規化のバックフィル。起動のたびに走るが冪等で、埋まっていれば no-op。
		// デプロイと手動 SQL 適用の順序に依存させないためここで行う
		// （DDL → バックフィル → listen の順が 1 プロセス内で保証される）。
		if err := BackfillUserNormalization(ctx, tx); err != nil {
			return err
		}
		if err := ApplyUserNormalizationConstraints(ctx, tx); err != nil {
			return err
		}
		if err := ApplyRichDocumentConstraints(ctx, tx); err != nil {
			return err
		}
		return ApplySessionNoteConstraints(ctx, tx)
	}); err != nil {
		return err
	}

	log.Println("migrate: knowledge base schema start")
	if err := ApplyKnowledgeBaseSchema(ctx, db); err != nil {
		return err
	}
	log.Println("migrate: knowledge base schema done")

	// テナント統合の Expand（companies → workspaces）。workspaces を参照する FK を張るため
	// ナレッジ基盤スキーマの後に置く。DDL もバックフィルも冪等で、埋まっていれば no-op。
	// 読み取りは引き続き company_id を見るので、この時点で挙動は何も変わらない。
	log.Println("migrate: tenant bridge start")
	if err := ApplyTenantBridgeSchema(ctx, db); err != nil {
		return err
	}
	if err := BackfillWorkspacesFromCompanies(ctx, db); err != nil {
		return err
	}
	log.Println("migrate: tenant bridge done")
	return nil
}

// withMigrateTx は advisory lock 付きの単一トランザクションで fn を実行する。
// 複数の ECS タスクが同時に起動しても DDL / seed / バックフィルが 1 つずつ流れるよう、
// 起動時マイグレーション共通のキーでトランザクションロックを取る
// （pgbouncer(transaction pooler) 前提のため、コミットで自動解放されないセッションロックは使わない）。
//
// ロック待ちは lock_timeout で必ず有限にし、超過したら指数バックオフで数回だけやり直してから
// 起動を失敗させる。
//
//   - やり直す理由: 長い書き込みトランザクションやローリングデプロイ中の先行タスクは、ふつう
//     数秒から十数秒で終わる一過性のもの。そこで即座に諦めるとデプロイが無駄に落ちる。
//     各段は冪等で、タイムアウトした試行はロールバック済みなので二重適用にならない。
//   - 最後は失敗させる理由: スキーマが半端なまま listen を始めるより、起動を落として ECS に
//     タスクを作り直させる方が安全。ローリングデプロイなら旧タスクが serving を続けるので、
//     外形的な停止にもならない。いちばん悪いのは無限に待って全ライターを詰まらせることなので、
//     待ち時間は必ず有限にする。
func withMigrateTx(ctx context.Context, db *sql.DB, label string, fn func(tx *sql.Tx) error) error {
	var lastErr error
	for attempt := 0; attempt <= migrateLockRetries; attempt++ {
		if attempt > 0 {
			wait := time.Duration(1<<(attempt-1)) * time.Second
			log.Printf("migrate: %s のロック待ちがタイムアウトしました。%s 後に再試行します（%d/%d）",
				label, wait, attempt, migrateLockRetries)
			if err := sleepCtx(ctx, wait); err != nil {
				return fmt.Errorf("%s: 再試行の待機が中断されました: %w", label, err)
			}
		}
		err := runMigrateTx(ctx, db, label, fn)
		if err == nil {
			return nil
		}
		if !isLockTimeout(err) {
			return err
		}
		lastErr = err
	}
	return fmt.Errorf(
		"%s: ロック待ちのタイムアウトが %d 回続いたため中断しました（長時間の書き込みトランザクションが残っていないか確認してください）: %w",
		label, migrateLockRetries+1, lastErr,
	)
}

// runMigrateTx は withMigrateTx の 1 回分の試行。
func runMigrateTx(ctx context.Context, db *sql.DB, label string, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("%s: トランザクション開始に失敗: %w", label, err)
	}
	defer func() { _ = tx.Rollback() }() // Commit 済みなら no-op

	// SET LOCAL はこのトランザクションの中だけに効く（接続をプールへ返しても残らない）。
	// パラメータプレースホルダを取れないので定数を埋め込む。
	if _, err := tx.ExecContext(ctx, `SET LOCAL lock_timeout = '`+migrateAdvisoryLockTimeout+`'`); err != nil {
		return fmt.Errorf("%s: lock_timeout の設定に失敗: %w", label, err)
	}
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock($1)`, migrateAdvisoryLockKey); err != nil {
		return fmt.Errorf("%s: advisory lock の取得に失敗: %w", label, err)
	}
	if _, err := tx.ExecContext(ctx, `SET LOCAL lock_timeout = '`+migrateLockTimeout+`'`); err != nil {
		return fmt.Errorf("%s: lock_timeout の設定に失敗: %w", label, err)
	}
	if err := fn(tx); err != nil {
		return fmt.Errorf("%s: %w", label, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("%s: コミットに失敗: %w", label, err)
	}
	return nil
}

// isLockTimeout は lock_timeout 超過（SQLSTATE 55P03）かを判定する。
func isLockTimeout(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == lockNotAvailableCode
}

// sleepCtx は ctx のキャンセルで打ち切れる待機。
func sleepCtx(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// ApplyRichDocumentConstraints は rich_documents の整合性制約を適用する（冪等）。
// schema/core.sql の CREATE TABLE にも同じ制約を書いてあるが、CREATE TABLE IF NOT EXISTS は
// 既に在るテーブルへ制約を足さないため、既存 DB に後から張る経路としてここを残す。
func ApplyRichDocumentConstraints(ctx context.Context, db Executor) error {
	stmts := []string{
		// owner_id → users.id。ユーザーの物理削除で文書も消す（論理削除運用なので通常は発火しない）。
		// 存在判定は conname だけでなく conrelid（テーブル）でも絞る。制約名は PostgreSQL では
		// テーブル単位でしか一意でないため、別テーブルに同名制約があっても取り違えないようにする。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_rich_documents_owner' AND conrelid = 'rich_documents'::regclass) THEN
				ALTER TABLE rich_documents
					ADD CONSTRAINT fk_rich_documents_owner
					FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE;
			END IF;
		END $$;`,
		// doc は tiptap のドキュメント JSON（object かつ type='doc'）に限る。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_rich_documents_doc' AND conrelid = 'rich_documents'::regclass) THEN
				ALTER TABLE rich_documents
					ADD CONSTRAINT ck_rich_documents_doc
					CHECK (jsonb_typeof(doc) = 'object' AND doc->>'type' = 'doc');
			END IF;
		END $$;`,
		// title 長の上限（アプリ側検証と二重の壁）。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_rich_documents_title_len' AND conrelid = 'rich_documents'::regclass) THEN
				ALTER TABLE rich_documents
					ADD CONSTRAINT ck_rich_documents_title_len
					CHECK (char_length(title) <= 200);
			END IF;
		END $$;`,
	}
	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

// ApplySessionNoteConstraints は session_notes に 1 セッション 1 ノートの一意制約を張る（冪等）。
// session_id には schema/core.sql が一意索引を張っているが、既存 DB では非一意のまま残っている
// ことがあるため、別名の一意インデックスを明示 SQL で必ず作る。
// 既存に session_id 重複があると作成に失敗するので、適用前に重複が無いことを確認すること。
//
// CREATE UNIQUE INDEX IF NOT EXISTS は索引が既に在ってスキップする場合でも session_notes の
// ShareLock を取り、トランザクションが終わるまで手放さない。起動のたびにノートの書き込みを
// 止めないよう、先にカタログを引いて未作成のときだけ発行する。
func ApplySessionNoteConstraints(ctx context.Context, db Executor) error {
	exists, err := indexExists(ctx, db, "uq_session_notes_session_id")
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	_, err = db.ExecContext(
		ctx,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_session_notes_session_id ON session_notes (session_id)`,
	)
	return err
}

// preRepairUsersForMigrate は正規化バックフィルの前提を満たすよう旧データを埋める（冪等）。
// users テーブル / 対象カラムが未作成の初回起動では no-op。
func preRepairUsersForMigrate(ctx context.Context, db Executor) error {
	_, err := db.ExecContext(ctx, `DO $$ BEGIN
		IF to_regclass('users') IS NOT NULL THEN
			UPDATE users SET email = '' WHERE email IS NULL;
			IF EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'users' AND column_name = 'role_id'
			) THEN
				UPDATE users SET role_id = 3 WHERE role_id IS NULL;
			END IF;
		END IF;
	END $$;`)
	return err
}

// SeedRoles はロールマスタを投入する（固定 ID・冪等）。起動時と結合テストのスキーマ構築で使う。
// 既存行は書き換えない（運用で説明文を直しても起動のたびに戻さない）。
func SeedRoles(ctx context.Context, db Executor) error {
	seeds := []domain.Role{
		{ID: domain.RoleIDSuperAdmin, Name: domain.RoleSuperAdmin, Description: "運営管理者"},
		{ID: domain.RoleIDCompanyAdmin, Name: domain.RoleCompanyAdmin, Description: "企業管理者"},
		{ID: domain.RoleIDTrainee, Name: domain.RoleTrainee, Description: "受講者"},
	}
	for _, r := range seeds {
		if _, err := db.ExecContext(
			ctx,
			`INSERT INTO roles (id, name, description, created_at, updated_at)
			 VALUES ($1, $2, $3, NOW(), NOW())
			 ON CONFLICT (id) DO NOTHING`,
			r.ID, string(r.Name), r.Description,
		); err != nil {
			return err
		}
	}
	return nil
}

// BackfillUserNormalization は正規化テーブル（users.role_id / user_oidc_identities）と
// email の正規形（users / invitations）の整合を起動時に保つ（冪等）。
// 正規化後は role_id が正で新コードが常に書くため、旧 role 文字列から
// role_id を「逆算」する同期は行わない（それをやると、role_id だけ更新した昇格を巻き戻す）。
// 旧カラム users.cognito_sub からの identity 補完だけは、旧カラム撤去（migrations/0021）の前後で
// 安全に流せるよう、カラムが存在する間のみカラム存在チェックでガードして実行する。
func BackfillUserNormalization(ctx context.Context, db Executor) error {
	// role_id が未設定の行は最小権限の trainee に倒す（読み出し側の LEFT JOIN で NULL role を作らない）。
	// role_id のみを触るため旧カラム撤去後も有効。
	if _, err := db.ExecContext(
		ctx,
		`UPDATE users SET role_id = $1 WHERE role_id IS NULL`, domain.RoleIDTrainee,
	); err != nil {
		return err
	}
	// email をアプリと同じ正規形（domain.NormalizeEmail = lower + 前後の EmailTrimCutset 除去）へ
	// 畳む（冪等。畳み済みの行は WHERE で外れるので no-op）。
	//
	// 索引と検索を正規形の式にするだけでも「空白付きの既存行が引けない」は解消するが、行の値が
	// 生のままだと、その環境では正規形の一意性を旧索引が守れない状態が残る（例: ' a@x.com' が
	// 在るところへアプリが 'a@x.com' を作れてしまい、同じ人の行が 2 つできる）。ここで先に畳んで
	// おくと、索引を張り替えられない環境でも「行の値 = 正規形」になり、旧索引がそのまま正規形の
	// 一意性を守る。
	//
	// 畳むと他のアクティブ行と衝突する行だけは触らない（別人かもしれない 2 行を勝手に 1 つの
	// アドレスへ寄せない）。残った衝突は ApplyUserNormalizationConstraints が WARNING で報告し、
	// FindActiveByEmail は複数件を曖昧としてログインを拒否する（fail closed）ので、
	// 解消は運用判断に委ねる。
	if _, err := db.ExecContext(
		ctx,
		`UPDATE users u
		    SET email = lower(btrim(u.email, E'\t\n\x0B\f\r '))
		  WHERE u.deleted_at IS NULL
		    AND u.email <> lower(btrim(u.email, E'\t\n\x0B\f\r '))
		    AND NOT EXISTS (
		        SELECT 1 FROM users o
		         WHERE o.id <> u.id
		           AND o.deleted_at IS NULL
		           AND lower(btrim(o.email, E'\t\n\x0B\f\r ')) = lower(btrim(u.email, E'\t\n\x0B\f\r '))
		    )`,
	); err != nil {
		return err
	}
	// 招待の email も同じ正規形へ畳む（保留中のみ）。ログイン時の招待ゲートは正規形の OIDC メールで
	// 引くため、大文字混じり・空白付きのまま残った pending 行は「招待したのに招待が見つからない」に
	// なる。invitations 側に一意制約は無いので衝突判定は要らない。
	if _, err := db.ExecContext(
		ctx,
		`UPDATE invitations
		    SET email = lower(btrim(email, E'\t\n\x0B\f\r '))
		  WHERE status = $1
		    AND email <> lower(btrim(email, E'\t\n\x0B\f\r '))`,
		domain.InvitationStatusPending,
	); err != nil {
		return err
	}
	// cognito_sub → user_oidc_identities（旧カラムが残っている間のみ・既存行はスキップ）。
	// 論理削除済みユーザーは対象外（identity が subject を占有すると再招待がログイン不能になる）。
	hasCognitoSub, err := columnExists(ctx, db, "users", "cognito_sub")
	if err != nil {
		return err
	}
	if hasCognitoSub {
		if _, err := db.ExecContext(
			ctx,
			`INSERT INTO user_oidc_identities (user_id, provider, subject, created_at, updated_at)
			 SELECT id, $1, cognito_sub, NOW(), NOW() FROM users
			 WHERE cognito_sub IS NOT NULL AND cognito_sub <> '' AND deleted_at IS NULL
			 ON CONFLICT DO NOTHING`, domain.OidcProviderCognito,
		); err != nil {
			return err
		}
	}
	// 論理削除済みユーザーに紐付く identity を掃除する（SoftDelete 側でも消すが、
	// 過去データと削除処理の失敗に対する自己修復として毎起動流す。冪等）。
	_, err = db.ExecContext(
		ctx,
		`DELETE FROM user_oidc_identities oi USING users u
		 WHERE oi.user_id = u.id AND u.deleted_at IS NOT NULL`,
	)
	return err
}

// columnExists は指定テーブルにカラムが存在するかを返す。旧カラム撤去の前後で
// バックフィルの分岐に使う（information_schema はトランザクション内でも現在のスキーマを見る）。
// columnExists は public スキーマの指定した表に、その列があるかを返す。
//
// table_schema = 'public' の絞りが要る理由:
//
//	この DB には Supabase 自身が持つ auth スキーマが同居していて、そこにも users という
//	表がある（auth.users・35 列）。スキーマを絞らずに information_schema を数えると、
//	アプリの public.users から列を消した後でも auth.users 側の同名列を拾って
//	「まだある」と誤判定する。実際 auth.users には role 列があり、これは
//	migrations/0021 が public.users から消す列と同じ名前。
//
//	誤判定すると「列があるはず」という前提で SQL を流し、実際には無いので
//	起動時マイグレーションが落ちる。アプリが上がらなくなるため、ここは必ず絞る。
func columnExists(ctx context.Context, db Executor, table, column string) (bool, error) {
	var n int64
	if err := db.QueryRowContext(
		ctx,
		`SELECT count(*) FROM information_schema.columns
		 WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2`, table, column,
	).Scan(&n); err != nil {
		return false, err
	}
	return n > 0, nil
}

// seedCompanies は既定の会社（自社）を投入する（固定 ID・冪等）。
func seedCompanies(ctx context.Context, db Executor) error {
	_, err := db.ExecContext(
		ctx,
		`INSERT INTO companies (id, name, created_at, updated_at)
		 VALUES ($1, $2, NOW(), NOW())
		 ON CONFLICT (id) DO NOTHING`,
		1, "株式会社FreStyle",
	)
	return err
}

// syncSeededSequences は seed 済みの表の採番列（シーケンス）を、実際に入っている最大 id へ合わせる。
//
// # なぜ必要か
//
// bigserial のような採番列は「INSERT で id を書かなければ nextval() が呼ばれ、シーケンスが
// 1 つ進む」という仕組みで動く。裏を返すと、id を明示して INSERT すると nextval() を通らず、
// シーケンスは 1 ミリも進まない。
//
// seedCompanies は固定 id（会社 1）を書いて投入する。そのため作られた直後の DB は
// 「行は入っているのにシーケンスは初期値のまま」になる。この状態で普通の新規作成
// （id を書かない INSERT）が来ると、nextval() が既に使われている id を返し、主キー衝突
// （SQLSTATE 23505）で落ちる。1 回落ちるとシーケンスは進むので 2 回目は成功する ——
// 「最初の 1 回だけ失敗する」という分かりにくい壊れ方になる。
//
// 実際、本番の companies_id_seq は last_value=1 / is_called=false のまま companies に
// id=1 が存在しており、会社を新規作成すると必ず初回が失敗する状態だった。
//
// # なぜ安全か
//
// setval は「次に配る番号」を動かすだけで、既存の行には一切触れない。同じ値を何度セットしても
// 結果は同じなので冪等で、起動のたびに走らせてよい。
//
// # roles を対象にしない理由
//
// roles.id は採番列ではない（core.sql で integer PRIMARY KEY・既定値なし。本番も同じ）。
// ロールは 1〜3 の固定マスタで、アプリから新規作成することが無いため採番列を持たせていない。
// 採番列が無い表に対して pg_get_serial_sequence は NULL を返し、setval(NULL, ...) は
// 何もせず NULL を返すだけなので害は無いが、意味の無い文は置かない。
//
// # 対象を増やすとき
//
// 「固定 id で seed していて、かつ採番列を持つ」表が増えたらここに 1 つ足す。テーブル名を
// 変数で組み立てず SQL に直接書くのは、識別子（テーブル名・列名）が $1 のような
// プレースホルダで渡せないため。値と違って識別子は SQL の構文そのものなので、
// 外部入力が混ざらない形に保つ。
func syncSeededSequences(ctx context.Context, db Executor) error {
	seeded := []struct {
		table string
		query string
	}{
		{
			table: "companies",
			// COALESCE の 1 は「行が 1 つも無いとき」の既定値。setval の第 2 引数は 1 以上で
			// なければならず、max(id) が NULL のまま渡すとエラーになるため。
			query: `SELECT setval(pg_get_serial_sequence('companies', 'id'), COALESCE((SELECT max(id) FROM companies), 1))`,
		},
	}
	for _, s := range seeded {
		// setval は値を返す関数なので SELECT で呼ぶ。返り値は使わない。
		if _, err := db.ExecContext(ctx, s.query); err != nil {
			return fmt.Errorf("%s の採番列の同期に失敗: %w", s.table, err)
		}
	}
	return nil
}

// ApplyUserNormalizationConstraints は正規化テーブルの整合性制約を適用する（冪等）。
// バックフィル後に呼ぶ前提（既存行が制約を満たしてから付ける）。FK / CHECK / 部分 UNIQUE は
// 既存データの状態に依存するため、CREATE TABLE ではなくここで明示 SQL として管理する。
func ApplyUserNormalizationConstraints(ctx context.Context, db Executor) error {
	stmts := []string{
		// roles.name: 空文字禁止。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_roles_name_not_empty') THEN
				ALTER TABLE roles ADD CONSTRAINT ck_roles_name_not_empty CHECK (name <> '');
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
		// users.email: アクティブ行（未論理削除）かつ正規形が非空に限った部分 UNIQUE。
		// 論理削除→同メール再招待と両立し、email claim の無い OIDC ユーザー（空文字）は対象外にする
		// （重複ガードと述語を必ず一致させること。ずれると起動失敗が自己修復しなくなる）。
		//
		// キーは email そのものではなく domain.NormalizeEmail と同じ正規形
		// lower(btrim(email, E'\t\n\x0B\f\r '))。アプリは畳んだ値を保存するが、索引の式が
		// アプリの正規形より緩いと「アプリでは同一・DB では別行」という食い違いが残り、大小文字
		// だけ・前後空白だけが違う 2 行が両方作れてしまう（正規化前に入った既存行も同じ穴になる）。
		// btrim の文字集合は domain.EmailTrimCutset と同じものを明示列挙する（btrim の既定は
		// 半角スペース 1 文字だけ、Go の strings.TrimSpace は Unicode 空白まで落とすため、
		// どちらの既定に寄せても両側は一致しない）。
		//
		// 既存データに重複がある場合は作成せず WARNING を出す（起動を落とさず、修正は運用判断に委ねる）。
		// 旧定義（生の email キー / lower(email) キー）からの張り替えは、畳んだ値に重複が
		// 無いことを確かめてから行う。先に落としてしまうと、新しい索引を作れない環境で保護が消えるため。
		// 空白だけが違う行は BackfillUserNormalization が先に畳んでいるので、ここに残る重複は
		// 「畳んでも別人か判断できない」本物の衝突だけになる。
		`DO $$ BEGIN
			IF EXISTS (
				SELECT 1 FROM pg_indexes WHERE indexname = 'uq_users_email_active'
				  AND indexdef NOT LIKE '%btrim%'
			) AND NOT EXISTS (
				SELECT 1 FROM users
				WHERE deleted_at IS NULL AND btrim(email, E'\t\n\x0B\f\r ') <> ''
				GROUP BY lower(btrim(email, E'\t\n\x0B\f\r ')) HAVING count(*) > 1
			) THEN
				DROP INDEX uq_users_email_active;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'uq_users_email_active') THEN
				IF EXISTS (
					SELECT 1 FROM users
					WHERE deleted_at IS NULL AND btrim(email, E'\t\n\x0B\f\r ') <> ''
					GROUP BY lower(btrim(email, E'\t\n\x0B\f\r ')) HAVING count(*) > 1
				) THEN
					RAISE WARNING 'users.email に（大小文字・前後空白を無視した）重複があるため uq_users_email_active を作成できません（重複を解消して再起動してください）';
				ELSE
					CREATE UNIQUE INDEX uq_users_email_active
						ON users (lower(btrim(email, E'\t\n\x0B\f\r ')))
						WHERE deleted_at IS NULL AND btrim(email, E'\t\n\x0B\f\r ') <> '';
				END IF;
			END IF;
		END $$;`,
	}
	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}
