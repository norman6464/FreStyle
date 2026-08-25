package database

import (
	"fmt"

	"github.com/norman6464/FreStyle/backend/internal/domain"
	"gorm.io/gorm"
)

// このファイルは運営権限の受け皿 users.is_platform_admin を用意する段（Expand）だけを担う。
// 適用直後は既存の運営管理者が全員 true になるので、挙動は何も変わらない。
//
// 列そのものは domain.User の GORM タグが持っており、まっさらな DB では AutoMigrate が作る。
// にもかかわらずここで明示 DDL を書くのは、バックフィルを「列が無かった 1 回だけ」に
// 縛るため。列の有無がその 1 回の印になる。
//
// 起動のたびに無条件で「role が super_admin なら true」を流してはいけない。それをやると、
// グループから外して false にした退任者が次のデプロイで復権する（role_id は下げないので
// 条件は永久に成立し続ける）。特に groups claim が載らない federated ユーザーは
// 失効経路が二度と走らないことがあり、復権したまま残る。

// ExpandUsersPlatformAdmin は users.is_platform_admin を追加し、既存の super_admin を
// true でバックフィルする（冪等・列が既にあれば no-op）。
// AutoMigrate より前に呼ぶこと（AutoMigrate が先に列を作ると、バックフィルの印が消える）。
func ExpandUsersPlatformAdmin(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 複数の ECS タスクが同時に起動しても DDL とバックフィルが 1 つずつ流れるよう、
		// 起動時マイグレーションと同じキーのトランザクションロックを取る
		// （pgbouncer 経由でも安全なようセッションロックは使わない）。
		if err := tx.Exec(`SELECT pg_advisory_xact_lock(?)`, migrateAdvisoryLockKey).Error; err != nil {
			return err
		}
		var usersTable *string
		if err := tx.Raw(`SELECT to_regclass('users')::text`).Scan(&usersTable).Error; err != nil {
			return err
		}
		if usersTable == nil {
			// まっさらな DB。users ごと AutoMigrate が作るので、埋める既存行も無い。
			return nil
		}
		has, err := columnExists(tx, "users", "is_platform_admin")
		if err != nil {
			return err
		}
		if has {
			return nil // 適用済み。ここで戻らないと退任者が復権する。
		}
		if err := tx.Exec(
			`ALTER TABLE users ADD COLUMN is_platform_admin boolean NOT NULL DEFAULT false`,
		).Error; err != nil {
			return err
		}
		// role_id は roles マスタの固定採番（domain.RoleIDSuperAdmin）。整数定数を
		// 文字列へ埋め込む（DO ブロックではないがプレースホルダを使わないのは、
		// 値の出どころがコード上の定数だけで外部入力を含まないため）。
		return tx.Exec(fmt.Sprintf(
			`UPDATE users SET is_platform_admin = true WHERE role_id = %d`,
			domain.RoleIDSuperAdmin,
		)).Error
	})
}
