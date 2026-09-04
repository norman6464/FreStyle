package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// BackfillWorkspacesFromCompanies が担うのは「会社を作ったら、対応するワークスペースを
// 必ず 1 つ用意する」という恒常的な同期。
//
// これはかつて company_id → workspace_id への一回きりの移行（Expand → Migrate → Contract の
// 3 段）の Expand 段だった。後半 2 段（各表への値の移送 / company_id 列の撤去）は既に完了し、
// company_id 列自体も schema.sql から撤去済みで、今この関数を再実行しても手を出す先が無い。
// Expand 段だけは移行専用ではなく、companies が存在し続ける限り恒常的に要る同期のため残す。

// workspaceSlugPrefix は自動採番した workspaces.slug の接頭辞。
const workspaceSlugPrefix = "ws-"

// BackfillWorkspacesFromCompanies は既存の会社に対応する workspaces 行を作り、
// companies.workspace_id を埋める（冪等）。
//
// 起動のたびに走るが、埋まっている行は対象から外れるので実質 no-op になる。
//
// is_active はここで写さない。停止の正本は workspaces.is_active に移り、companies 側を
// 変える手段（会社の有効/無効 API）はもう無い。それでも毎起動 companies から写すと、
// ワークスペースを停止しても次の起動で会社側の値に巻き戻され、停止そのものが効かなくなる。
func BackfillWorkspacesFromCompanies(ctx context.Context, db *sql.DB) error {
	return withMigrateTx(ctx, db, "会社→ワークスペースのバックフィル", func(tx *sql.Tx) error {
		return createWorkspacesForCompanies(ctx, tx)
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
