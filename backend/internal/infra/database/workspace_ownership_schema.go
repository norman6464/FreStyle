package database

import (
	"context"
	"database/sql"
	"fmt"
)

// workspaces に「個人ワークスペースの持ち主」を表す列と制約を足す（tenant_bridge.go と同じ
// 理由・同じ ALTER TABLE ADD COLUMN IF NOT EXISTS の作法で、既存本番表へ列を届ける）。

// workspaceOwnershipSchemaStatements は Expand で足す列と制約（冪等）。
var workspaceOwnershipSchemaStatements = []string{
	`DO $$ BEGIN
		IF NOT EXISTS (
			SELECT 1 FROM information_schema.columns
			 WHERE table_schema = current_schema()
			   AND table_name = 'workspaces' AND column_name = 'personal_owner_user_id'
		) THEN
			ALTER TABLE workspaces ADD COLUMN personal_owner_user_id bigint;
		END IF;
	END $$;`,
	// 作った人を物理削除しても中身は消さない。持ち主のいない箱として残り、
	// 招かれた他のメンバーはそのまま使い続けられる。
	`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_workspaces_personal_owner' AND conrelid = 'workspaces'::regclass) THEN
			ALTER TABLE workspaces
				ADD CONSTRAINT fk_workspaces_personal_owner
				FOREIGN KEY (personal_owner_user_id) REFERENCES users (id) ON DELETE SET NULL;
		END IF;
	END $$;`,
	// 1 人につき個人ワークスペースは 1 つ。サインアップの再送・並行実行でも 2 つ目が作れない
	// （check-then-act をアプリに書かずに済む。ON CONFLICT の推論先にもなる）。
	`DO $$ BEGIN
		IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'uq_workspaces_personal_owner') THEN
			CREATE UNIQUE INDEX uq_workspaces_personal_owner
				ON workspaces (personal_owner_user_id) WHERE personal_owner_user_id IS NOT NULL;
		END IF;
	END $$;`,
}

// ApplyWorkspaceOwnershipSchema は workspaces に personal_owner_user_id 列と制約を足す（冪等）。
// users を参照する FK を張るため中核スキーマの後、workspaces を参照するため
// ApplyKnowledgeBaseSchema の後に呼ぶこと。
func ApplyWorkspaceOwnershipSchema(ctx context.Context, db *sql.DB) error {
	return withMigrateTx(ctx, db, "個人ワークスペース所有者スキーマ", func(tx *sql.Tx) error {
		for _, stmt := range workspaceOwnershipSchemaStatements {
			if _, err := tx.ExecContext(ctx, stmt); err != nil {
				return fmt.Errorf("DDL の適用に失敗: %w", err)
			}
		}
		return nil
	})
}
