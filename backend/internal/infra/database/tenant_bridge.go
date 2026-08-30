package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// companies を workspaces へ畳む移行を担う。段は 3 つで、Migrate() がこの順に呼ぶ:
//
//	Expand   … schema.sql（節Ⅳ・Ⅴ）が workspace_id 列と FK を足す
//	Migrate  … MigrateWorkspaceIDsFromCompanyID が company_id から値を写す
//	Contract … DropCompanyIDColumns が company_id 列を落とす
//
// **順序を崩さないこと。** Contract を先に走らせると company_id → workspace_id の対応が
// 失われ、既存のユーザー・教材・招待がすべて所属不明になる（復元できない）。
//
// 移行が済んだ後の起動では、Migrate 段は「company_id 列がもう無い」ことを見て何もせず、
// Contract 段も同じ判定で no-op になる。残るのは companies↔workspaces の 1:1 紐付けと
// companies.is_active → workspaces.is_active の反映だけ（正本は companies）。

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

// queryRower は *sql.DB と *sql.Tx の両方を受けるための最小インターフェース。
type queryRower interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// companyIDColumnExists は指定テーブルに company_id 列がまだ在るかを返す。
//
// 移行の段（Expand → Migrate → Contract）を跨いで冪等にするための判定。列を落とした後の
// 起動では、移送も DROP もこの判定で早期 return して何もしない。
func companyIDColumnExists(ctx context.Context, q queryRower, table string) (bool, error) {
	var exists bool
	if err := q.QueryRowContext(
		ctx,
		`SELECT EXISTS (
		   SELECT 1 FROM information_schema.columns
		    WHERE table_schema = current_schema()
		      AND table_name = $1 AND column_name = 'company_id'
		 )`,
		table,
	).Scan(&exists); err != nil {
		return false, fmt.Errorf("%s.company_id の有無の確認に失敗: %w", table, err)
	}
	return exists, nil
}

// MigrateWorkspaceIDsFromCompanyID は company_id から workspace_id への一度きりの移送。
//
// workspace_id 列は schema.sql（節Ⅳ・Ⅴ）が用意するが、既存行の値は空のままなので、
// company_id がまだ在るあいだにここで写す。**DropCompanyIDColumns より先に必ず呼ぶこと。**
//
// 判定は表ごとに行う。まっさらな DB では schema.sql が company_id を持たない表を作るので
// 全表で no-op になり、移行中の DB では残っている表だけが対象になる（表ごとに移行の
// 進み方が違う状態でも正しく動く）。
func MigrateWorkspaceIDsFromCompanyID(ctx context.Context, db *sql.DB) error {
	return withMigrateTx(ctx, db, "company_id から workspace_id への移送", func(tx *sql.Tx) error {
		for _, m := range workspaceIDMirrors {
			exists, err := companyIDColumnExists(ctx, tx, m.table)
			if err != nil {
				return err
			}
			if !exists {
				continue // その表は移送済み（列がもう無い）
			}
			if err := m.mirror(ctx, tx); err != nil {
				return err
			}
		}
		return nil
	})
}

// workspaceIDMirror は 1 表ぶんの移送。table は列の有無を確かめるための表名で、
// mirror が実際に値を写す。
type workspaceIDMirror struct {
	table  string
	mirror func(context.Context, *sql.Tx) error
}

// workspaceIDMirrors は company_id を持つ各表から workspace_id へ値を写す関数。
//
// テーブル名を文字列結合で SQL に埋め込まず、テーブルごとに固定文字列のクエリを持つ
// （動的な SQL 生成は避ける。CLAUDE.md の生 SQL 規約と gosec の両方に沿う）。
//
// company_id が NULL の行（未所属のユーザー、会社を持たない文書）は JOIN 条件に一致しないため
// 自然に対象から外れ、workspace_id も NULL のまま残る。未所属は「どの会社でもない」であって、
// 既定のテナントへ流し込んでよいものではない。
var workspaceIDMirrors = []workspaceIDMirror{
	{"users", mirrorUsersWorkspaceFromCompany},
	{"courses", mirrorCoursesWorkspaceFromCompany},
	{"course_chapters", mirrorCourseChaptersWorkspaceFromCompany},
	{"company_exercises", mirrorCompanyExercisesWorkspaceFromCompany},
	{"invitations", mirrorInvitationsWorkspaceFromCompany},
	{"rich_documents", mirrorRichDocumentsWorkspaceFromCompany},
}

// mirrorUsersWorkspaceFromCompany は所属会社のワークスペースをユーザーへ写す。
//
// updated_at は触らない。所属という同じ事実の写しを別の列へ置くだけの移行処理で、
// 利用者から見た更新ではないため（API に出る更新日時を動かさない）。
func mirrorUsersWorkspaceFromCompany(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE users u SET workspace_id = c.workspace_id
		 FROM companies c
		 WHERE u.company_id = c.id
		   AND c.workspace_id IS NOT NULL
		   AND u.workspace_id IS DISTINCT FROM c.workspace_id`,
	); err != nil {
		return fmt.Errorf("users へのワークスペース反映に失敗: %w", err)
	}
	return nil
}

func mirrorCoursesWorkspaceFromCompany(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE courses t SET workspace_id = c.workspace_id
		 FROM companies c
		 WHERE t.company_id = c.id
		   AND c.workspace_id IS NOT NULL
		   AND t.workspace_id IS DISTINCT FROM c.workspace_id`,
	); err != nil {
		return fmt.Errorf("courses へのワークスペース反映に失敗: %w", err)
	}
	return nil
}

func mirrorCourseChaptersWorkspaceFromCompany(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE course_chapters t SET workspace_id = c.workspace_id
		 FROM companies c
		 WHERE t.company_id = c.id
		   AND c.workspace_id IS NOT NULL
		   AND t.workspace_id IS DISTINCT FROM c.workspace_id`,
	); err != nil {
		return fmt.Errorf("course_chapters へのワークスペース反映に失敗: %w", err)
	}
	return nil
}

func mirrorCompanyExercisesWorkspaceFromCompany(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE company_exercises t SET workspace_id = c.workspace_id
		 FROM companies c
		 WHERE t.company_id = c.id
		   AND c.workspace_id IS NOT NULL
		   AND t.workspace_id IS DISTINCT FROM c.workspace_id`,
	); err != nil {
		return fmt.Errorf("company_exercises へのワークスペース反映に失敗: %w", err)
	}
	return nil
}

func mirrorInvitationsWorkspaceFromCompany(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE invitations t SET workspace_id = c.workspace_id
		 FROM companies c
		 WHERE t.company_id = c.id
		   AND c.workspace_id IS NOT NULL
		   AND t.workspace_id IS DISTINCT FROM c.workspace_id`,
	); err != nil {
		return fmt.Errorf("invitations へのワークスペース反映に失敗: %w", err)
	}
	return nil
}

func mirrorRichDocumentsWorkspaceFromCompany(ctx context.Context, tx *sql.Tx) error {
	if _, err := tx.ExecContext(
		ctx,
		`UPDATE rich_documents t SET workspace_id = c.workspace_id
		 FROM companies c
		 WHERE t.company_id = c.id
		   AND c.workspace_id IS NOT NULL
		   AND t.workspace_id IS DISTINCT FROM c.workspace_id`,
	); err != nil {
		return fmt.Errorf("rich_documents へのワークスペース反映に失敗: %w", err)
	}
	return nil
}

// DropCompanyIDColumns は所属参照としての company_id を落とす（移行の Contract 段）。
//
// **MigrateWorkspaceIDsFromCompanyID の後に必ず呼ぶこと。** 先に落とすと company_id →
// workspace_id の対応が失われ、既存の行がすべて所属不明になる。
//
// companies テーブル自体は残る。落とすのは「テナント参照としての company_id」だけで、
// 会社という実体は SuperAdmin の管理対象として引き続き必要。
//
// 列が既に無ければ何もしない（冪等）。カタログを見てから ALTER するのは、素の ALTER TABLE が
// 列が無くてスキップする場合でも先に ACCESS EXCLUSIVE ロックを取り、起動のたびにその表を
// 止めてしまうため。依存するインデックスは DROP COLUMN が一緒に落とす。
func DropCompanyIDColumns(ctx context.Context, db *sql.DB) error {
	return withMigrateTx(ctx, db, "company_id 列の撤去", func(tx *sql.Tx) error {
		for _, drop := range companyIDDrops {
			if err := drop(ctx, tx); err != nil {
				return err
			}
		}
		return nil
	})
}

// companyIDDrops は表ごとの DROP COLUMN。ミラーと同じ理由でテーブル名を SQL へ埋め込まない。
var companyIDDrops = []func(context.Context, *sql.Tx) error{
	dropUsersCompanyID,
	dropCoursesCompanyID,
	dropCourseChaptersCompanyID,
	dropCompanyExercisesCompanyID,
	dropInvitationsCompanyID,
	dropRichDocumentsCompanyID,
}

func dropUsersCompanyID(ctx context.Context, tx *sql.Tx) error {
	return dropCompanyIDIfPresent(ctx, tx, "users", `ALTER TABLE users DROP COLUMN company_id`)
}

func dropCoursesCompanyID(ctx context.Context, tx *sql.Tx) error {
	return dropCompanyIDIfPresent(ctx, tx, "courses", `ALTER TABLE courses DROP COLUMN company_id`)
}

func dropCourseChaptersCompanyID(ctx context.Context, tx *sql.Tx) error {
	return dropCompanyIDIfPresent(ctx, tx, "course_chapters", `ALTER TABLE course_chapters DROP COLUMN company_id`)
}

func dropCompanyExercisesCompanyID(ctx context.Context, tx *sql.Tx) error {
	return dropCompanyIDIfPresent(ctx, tx, "company_exercises", `ALTER TABLE company_exercises DROP COLUMN company_id`)
}

func dropInvitationsCompanyID(ctx context.Context, tx *sql.Tx) error {
	return dropCompanyIDIfPresent(ctx, tx, "invitations", `ALTER TABLE invitations DROP COLUMN company_id`)
}

func dropRichDocumentsCompanyID(ctx context.Context, tx *sql.Tx) error {
	return dropCompanyIDIfPresent(ctx, tx, "rich_documents", `ALTER TABLE rich_documents DROP COLUMN company_id`)
}

// dropCompanyIDIfPresent は列が在るときだけ dropSQL を流す。dropSQL は呼び出し側が持つ
// 固定文字列のみ（外部入力を SQL に組み込まない）。
func dropCompanyIDIfPresent(ctx context.Context, tx *sql.Tx, table, dropSQL string) error {
	exists, err := companyIDColumnExists(ctx, tx, table)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	if _, err := tx.ExecContext(ctx, dropSQL); err != nil {
		return fmt.Errorf("%s.company_id の削除に失敗: %w", table, err)
	}
	return nil
}
