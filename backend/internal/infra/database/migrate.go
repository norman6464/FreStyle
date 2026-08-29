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

// Migrate は起動時にスキーマを適用する。schema/schema.sql（中核 → ノート → 権限 →
// テナント橋渡しの順に並ぶ）が正本で、バックフィル・制約もその中の DO ブロックとして
// 埋め込まれている。RESET_DB=true のときは public schema を完全 wipe してから再構築する
// （一回限りの初期構築用）。
//
// roles マスタは users.role_id が参照する FK 先なので投入まで行う（SeedRoles は
// ON CONFLICT DO NOTHING で冪等・既存行は書き換えない）。新規環境・DR 復元で
// 真っさらな DB に対して起動しても、ユーザー作成・招待受諾が初回から通る。
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
	log.Println("migrate: core schema start")
	if err := ApplyCoreSchema(ctx, db); err != nil {
		return err
	}
	log.Println("migrate: core schema done")
	if err := withMigrateTx(ctx, db, "ロールマスタの投入", func(tx *sql.Tx) error {
		return SeedRoles(ctx, tx)
	}); err != nil {
		return err
	}
	// 演習データ(PHP / Go / Docker / Linux / Git など)は問題文・期待出力を公開リポに露出させない
	// ため本体には埋め込まず、非公開の教材リポ(frestyle-teaching-materials/exercises/<lang>/*.md)を
	// 唯一の正本とし、seed.py が生成する UPSERT SQL を Supabase に流して投入する。
	log.Println("migrate: knowledge base schema start")
	if err := ApplyKnowledgeBaseSchema(ctx, db); err != nil {
		return err
	}
	log.Println("migrate: knowledge base schema done")

	// テナント統合の Expand（companies → workspaces）。DDL は schema.sql（Ⅳ）で
	// 済んでいるので、ここは既存の会社に対応する workspaces 行を作るバックフィルだけ。
	log.Println("migrate: tenant bridge start")
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

// SeedRoles はロールマスタを投入する（固定 ID・冪等）。起動時の Migrate が呼ぶ他、
// 結合テストのスキーマ構築（testsupport.OpenTestDB）も明示的に呼ぶ。
// 既存行は書き換えない（運用で説明文を直しても呼び直すたびに戻さない）。
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
