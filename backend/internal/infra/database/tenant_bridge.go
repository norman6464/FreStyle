package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// companies から workspaces への 1:1 紐付けを用意する（会社が作られたら、対応する
// ワークスペースを冪等に用意する）。company_id は撤去済みで、業務テーブルの所属参照は
// すべて workspace_id を直接指定する書き込み経路に一本化されている（このファイル名の
// 「橋渡し」は現在は companies↔workspaces の紐付けだけを指す）。
//
// companies.is_active → workspaces.is_active の反映も併せて行う（正本は companies）。

// workspaceSlugPrefix は自動採番した workspaces.slug の接頭辞。
const workspaceSlugPrefix = "ws-"

// BackfillWorkspacesFromCompanies は既存の会社に対応する workspaces 行を作り、
// companies.workspace_id を埋め、is_active をワークスペースへ反映する（冪等）。
//
// 起動のたびに走るが、埋まっている行は対象から外れるので実質 no-op になる。
func BackfillWorkspacesFromCompanies(ctx context.Context, db *sql.DB) error {
	return withMigrateTx(ctx, db, "会社→ワークスペースのバックフィル", func(tx *sql.Tx) error {
		if err := createWorkspacesForCompanies(ctx, tx); err != nil {
			return err
		}
		if err := mirrorCompanySettingsToWorkspaces(ctx, tx); err != nil {
			return err
		}
		return nil
	})
}

// createWorkspacesForCompanies は workspace_id が未設定の会社ごとに workspaces 行を作り、
// 会社側からその行を指す。1 会社ずつ「作る → 指す」を同じトランザクションで行うため、
// 行を作ったのに指せていない中途半端な状態は残らない。
func createWorkspacesForCompanies(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx,
		`SELECT id, name, is_active
		 FROM companies WHERE workspace_id IS NULL ORDER BY id`)
	if err != nil {
		return fmt.Errorf("未対応の会社の取得に失敗: %w", err)
	}
	type companyRow struct {
		id       int64
		name     string
		isActive bool
	}
	var targets []companyRow
	for rows.Next() {
		var c companyRow
		if err := rows.Scan(&c.id, &c.name, &c.isActive); err != nil {
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
			`INSERT INTO workspaces (id, slug, name, is_active)
			 VALUES ($1, $2, $3, $4)`,
			id, workspaceSlugFor(id), name, c.isActive,
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
// 移行中の正本は companies 側なので、食い違ったら companies に合わせるのが常に正しい。
// 一致していれば 0 件更新で、updated_at も動かない。
func mirrorCompanySettingsToWorkspaces(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE workspaces w
		 SET is_active = c.is_active,
		     updated_at = now()
		 FROM companies c
		 WHERE c.workspace_id = w.id
		   AND w.is_active IS DISTINCT FROM c.is_active`,
	); err != nil {
		return fmt.Errorf("ワークスペースへの設定反映に失敗: %w", err)
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
