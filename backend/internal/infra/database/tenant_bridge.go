package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// このファイルは「テナントの正本を companies から workspaces へ移す」移行のうち、
// 事実を両側に持たせる段（Expand）だけを担う。読み取りは 1 つも変えない。
//
// なぜ 2 つのテナント表現が並ぶのか:
//   - companies は現行アプリのテナント。users.company_id が指すが、FK は 1 本も無く
//     境界はアプリの if だけが守っている。行を増やす経路も本番コードに存在しない。
//   - workspaces はナレッジ基盤のテナント。配下は複合 FK で DB が境界を守っている。
//
// 最終的に companies は畳んで消し、workspaces を唯一のテナントにする。
// ただし本番は ECS のローリングデプロイで新旧タスクが同時に走る瞬間があるため、
// 旧タスクが読み書きする列を先に消すと落ちる。そこで
// 「両方に書く（この段） → 読みを移す → 旧列を消す」の順で進める。
//
// したがって companies.workspace_id は恒久的な 1:1 の関連ではなく、移行期間だけの橋渡しで、
// companies を畳むときに列ごと消える。users.workspace_id だけが残って所属の正本になる。

// workspaceSlugPrefix は自動採番した workspaces.slug の接頭辞。
const workspaceSlugPrefix = "ws-"

// tenantBridgeSchemaStatements は Expand で足す列と制約（冪等）。
//
// companies / users は schema/core.sql が作るテーブルだが、この 2 列だけは CREATE TABLE ではなく
// ALTER TABLE ADD COLUMN IF NOT EXISTS で足す。CREATE TABLE IF NOT EXISTS は既に在るテーブルへ
// 列を追加しないため、既に本番にあるテーブルへ列を届ける経路がこれしかないから。
//
// 列は必ずテーブルの末尾に付く（ALTER TABLE ADD COLUMN の挙動）。schema/core.sql でも
// 最後に書いてあることが前提で、ずれると SELECT * の詰め替えが位置ずれで壊れる。
//
// ADD COLUMN IF NOT EXISTS を素で流さず、カタログを見て未作成のときだけ ALTER する。
// ALTER TABLE は列が既に在ってスキップする場合でも先に AccessExclusiveLock を取り、
// トランザクションが終わるまで手放さない（読み取りまで止まる）。列が出揃っている通常の起動で
// companies / users を掴まないよう、事前チェックで ALTER 自体を出さない。
var tenantBridgeSchemaStatements = []string{
	addWorkspaceIDColumnStatement("companies"),
	addWorkspaceIDColumnStatement("users"),
	// 会社とワークスペースは 1:1。移行中に 2 つの会社が同じワークスペースを指す状態を作らない。
	`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'uq_companies_workspace_id') THEN
			CREATE UNIQUE INDEX uq_companies_workspace_id ON companies (workspace_id) WHERE workspace_id IS NOT NULL;
		END IF;
	END $$;`,
	// 存在しないワークスペースを指せないようにする（company_id には FK が無く、
	// 同じ轍を踏まないために新しい列には最初から DB の壁を立てる）。
	// 参照されている workspaces の行は消せない（既定の NO ACTION）。テナントの実体を
	// 消す操作は所属の付け替えを伴うべきで、黙って道連れにしてよいものではない。
	`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_companies_workspace' AND conrelid = 'companies'::regclass) THEN
			ALTER TABLE companies
				ADD CONSTRAINT fk_companies_workspace
				FOREIGN KEY (workspace_id) REFERENCES workspaces (id);
		END IF;
	END $$;`,
	`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_users_workspace' AND conrelid = 'users'::regclass) THEN
			ALTER TABLE users
				ADD CONSTRAINT fk_users_workspace
				FOREIGN KEY (workspace_id) REFERENCES workspaces (id);
		END IF;
	END $$;`,
}

// addWorkspaceIDColumnStatement は workspace_id 列を「無ければ足す」DO ブロックを組み立てる。
// テーブル名は呼び出し側のリテラルだけを渡す（外部入力は来ない）。
func addWorkspaceIDColumnStatement(table string) string {
	return `DO $$ BEGIN
		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			 WHERE table_schema = current_schema()
			   AND table_name = '` + table + `' AND column_name = 'workspace_id'
		) THEN
			ALTER TABLE ` + table + ` ADD COLUMN workspace_id uuid;
		END IF;
	END $$;`
}

// ApplyTenantBridgeSchema は companies / users に workspace_id 列と制約を足す（冪等）。
// workspaces を参照する FK を張るため、ApplyKnowledgeBaseSchema の後に呼ぶこと。
func ApplyTenantBridgeSchema(ctx context.Context, db *sql.DB) error {
	return withMigrateTx(ctx, db, "テナント橋渡しスキーマ", func(tx *sql.Tx) error {
		for _, stmt := range tenantBridgeSchemaStatements {
			if _, err := tx.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("DDL の適用に失敗: %w", err)
			}
		}
		return nil
	})
}

// BackfillWorkspacesFromCompanies は既存の会社に対応する workspaces 行を作り、
// companies.workspace_id / users.workspace_id を埋める（冪等）。
//
// 起動のたびに走るが、埋まっている行は対象から外れるので実質 no-op になる。
// 途中まで進んだ状態から再開しても矛盾しないよう、各段階の WHERE を
// 「まだ埋まっていない行だけ」に絞ってある。
func BackfillWorkspacesFromCompanies(ctx context.Context, db *sql.DB) error {
	return withMigrateTx(ctx, db, "会社→ワークスペースのバックフィル", func(tx *sql.Tx) error {
		if err := createWorkspacesForCompanies(ctx, tx); err != nil {
			return err
		}
		if err := mirrorCompanySettingsToWorkspaces(ctx, tx); err != nil {
			return err
		}
		return mirrorUserWorkspaceFromCompany(ctx, tx)
	})
}

// createWorkspacesForCompanies は workspace_id が未設定の会社ごとに workspaces 行を作り、
// 会社側からその行を指す。1 会社ずつ「作る → 指す」を同じトランザクションで行うため、
// 行を作ったのに指せていない中途半端な状態は残らない。
func createWorkspacesForCompanies(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx,
		`SELECT id, name, ai_chat_enabled_for_trainees, is_active
		 FROM companies WHERE workspace_id IS NULL ORDER BY id`)
	if err != nil {
		return fmt.Errorf("未対応の会社の取得に失敗: %w", err)
	}
	type companyRow struct {
		id           int64
		name         string
		aiChatForAll bool
		isActive     bool
	}
	var targets []companyRow
	for rows.Next() {
		var c companyRow
		if err := rows.Scan(&c.id, &c.name, &c.aiChatForAll, &c.isActive); err != nil {
			_ = rows.Close()
			return fmt.Errorf("会社行の読み取りに失敗: %w", err)
		}
		targets = append(targets, c)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return fmt.Errorf("会社行の走査に失敗: %w", err)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("会社行のクローズに失敗: %w", err)
	}

	for _, c := range targets {
		id, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("ワークスペース ID の採番に失敗: %w", err)
		}
		// 表示名は会社名をそのまま写す（workspaces.name は varchar(200)）。
		name := truncateRunes(c.name, 200)
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO workspaces (id, slug, name, ai_chat_enabled_for_trainees, is_active)
			 VALUES ($1, $2, $3, $4, $5)`,
			id, workspaceSlugFor(id), name, c.aiChatForAll, c.isActive,
		); err != nil {
			return fmt.Errorf("ワークスペースの作成に失敗（company=%d）: %w", c.id, err)
		}
		if _, err := tx.ExecContext(
			ctx,
			`UPDATE companies SET workspace_id = $1 WHERE id = $2 AND workspace_id IS NULL`,
			id, c.id,
		); err != nil {
			return fmt.Errorf("会社へのワークスペース紐付けに失敗（company=%d）: %w", c.id, err)
		}
	}
	return nil
}

// mirrorCompanySettingsToWorkspaces は会社のテナント設定をワークスペースへ写す。
// 書き込み経路（company repository）でも同じ写しを行うが、ここでも毎起動ずれを直す。
// 移行中の正本は companies 側なので、食い違ったら companies に合わせるのが常に正しい
// （正本が workspaces へ移る段で、この向きは逆転させて撤去する）。
// 一致していれば 0 件更新で、updated_at も動かない。
func mirrorCompanySettingsToWorkspaces(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE workspaces w
		 SET ai_chat_enabled_for_trainees = c.ai_chat_enabled_for_trainees,
		     is_active = c.is_active,
		     updated_at = now()
		 FROM companies c
		 WHERE c.workspace_id = w.id
		   AND (w.ai_chat_enabled_for_trainees IS DISTINCT FROM c.ai_chat_enabled_for_trainees
		        OR w.is_active IS DISTINCT FROM c.is_active)`,
	); err != nil {
		return fmt.Errorf("ワークスペースへの設定反映に失敗: %w", err)
	}
	return nil
}

// mirrorUserWorkspaceFromCompany は所属会社のワークスペースをユーザーへ写す。
// 会社に属さないユーザー（company_id IS NULL）は NULL のまま残す — 未所属は
// 「どの会社でもない」であって、既定のテナントへ流し込んでよいものではない。
//
// updated_at は触らない。この列は所属という同じ事実の写しを別の場所へ置くだけの
// 移行処理で、利用者から見た更新ではないため（API に出る更新日時を動かさない）。
func mirrorUserWorkspaceFromCompany(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE users u SET workspace_id = c.workspace_id
		 FROM companies c
		 WHERE u.company_id = c.id
		   AND c.workspace_id IS NOT NULL
		   AND u.workspace_id IS DISTINCT FROM c.workspace_id`,
	); err != nil {
		return fmt.Errorf("ユーザーへのワークスペース反映に失敗: %w", err)
	}
	return nil
}

// workspaceSlugFor はワークスペース自身の ID から slug を採番する。
//
// slug はグローバル一意・1..64 文字・URL に出る識別子という制約がある。会社名は使えない:
// companies.name には一意制約が無く、日本語も入るので、一意にも URL 安全にもならない。
// 会社 ID を混ぜるのも避ける — companies はいずれ消える概念で、消えたあとに
// 意味の通らない名前が残る。ワークスペース自身の ID なら、companies が無くなっても
// 「そのワークスペースの識別子」として意味が通り、衝突もしない（UUID の一意性そのもの）。
//
// これは人が読むための名前ではなく、機械が付ける初期値。人が読む slug が要るなら
// 後から付け替えられる（slug は UNIQUE なだけの普通の列）。
func workspaceSlugFor(id uuid.UUID) string {
	return workspaceSlugPrefix + strings.ReplaceAll(id.String(), "-", "")
}

// truncateRunes は文字数（rune 数）で切り詰める。varchar(n) の n は文字数なので、
// バイト数で切るとマルチバイトの会社名で長さ判定がずれる。
func truncateRunes(s string, limit int) string {
	r := []rune(s)
	if len(r) <= limit {
		return s
	}
	return string(r[:limit])
}
