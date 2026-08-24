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
		// ナレッジ基盤（本文をブロック行に分解して持つ新スキーマ）。
		// FK / CHECK / 部分 UNIQUE は ApplyKnowledgeBaseConstraints が張る。
		&domain.Workspace{},
		&domain.Space{},
		&domain.Page{},
		&domain.Block{},
		&domain.PagePath{},
		&domain.PageSnapshot{},
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
		if err := ApplyRichDocumentConstraints(tx); err != nil {
			return err
		}
		return ApplyKnowledgeBaseConstraints(tx)
	})
}

// ApplyRichDocumentConstraints は rich_documents の整合性制約を適用する（冪等）。
// GORM の AutoMigrate は FK / CHECK を表現できないため、ここで明示 SQL として管理する。
func ApplyRichDocumentConstraints(db *gorm.DB) error {
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
		if err := db.Exec(stmt).Error; err != nil {
			return err
		}
	}
	return nil
}

// ApplyKnowledgeBaseConstraints はナレッジ基盤（workspaces / spaces / pages / blocks /
// page_paths / page_snapshots）の整合性制約を適用する（冪等）。
// GORM の AutoMigrate は複合 FK / CHECK / 部分 UNIQUE を表現できないため、ここで明示 SQL として管理する。
//
// 設計の柱は 2 つ:
//
//	(1) 境界越えを DB で塞ぐ。親子の FK は必ず「入れ物」の列を含む複合 FK にし、
//	    別のテナント / スペース / ページの行を親にできないようにする。
//	    木はそれぞれの入れ物の中で閉じる: ページの木はスペースの中、ブロックの木はページの中。
//	    入れ物をまたぐ親子を許すと、入れ物を消したときに ON DELETE CASCADE が
//	    別の入れ物に残るはずの行まで道連れにする。
//	    そのために参照先へ (workspace_id, …, id) の複合 UNIQUE を先に張る。
//	(2) 並び順は分数インデックス（internal/pkg/fracindex）の文字列キー。
//	    同じ親の中で position が重複しないことを UNIQUE で守り、既定値は置かない（採番はアプリ側）。
func ApplyKnowledgeBaseConstraints(db *gorm.DB) error {
	stmts := []string{
		// position 列のコレーションを "C"（バイト順）に固定する。
		// 分数インデックスは「文字列の辞書順 = 並び順」が前提で、Go 側はバイト比較で判断する。
		// DB の既定がロケール依存のコレーション（例: en_US.utf8）だと 'a' < 'B' のように並び、
		// ORDER BY position がアプリの認識とずれる。ここで揃えておく。
		// 既定コレーションの列は information_schema 上 collation_name が NULL になるため、
		// 'C' でない（NULL を含む）ときだけ ALTER する。
		`DO $$ BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = current_schema() AND table_name = 'pages'
				  AND column_name = 'position' AND collation_name IS DISTINCT FROM 'C'
			) THEN
				ALTER TABLE pages ALTER COLUMN "position" TYPE text COLLATE "C";
			END IF;
		END $$;`,
		`DO $$ BEGIN
			IF EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = current_schema() AND table_name = 'blocks'
				  AND column_name = 'position' AND collation_name IS DISTINCT FROM 'C'
			) THEN
				ALTER TABLE blocks ALTER COLUMN "position" TYPE text COLLATE "C";
			END IF;
		END $$;`,

		// --- UNIQUE ---
		// UNIQUE は「制約(ADD CONSTRAINT)」ではなく「UNIQUE 索引」で張る。GORM の AutoMigrate は
		// information_schema の UNIQUE 制約を見て「タグに unique が無い列に UNIQUE 制約が付いている」と
		// 判断すると、自分の命名規則の制約名(uni_<table>_<column>)を DROP しようとして毎起動落ちる
		// （複合 UNIQUE でも構成列それぞれが unique 扱いになるため同じことが起きる）。
		// UNIQUE 索引は AutoMigrate の関知外なので、この衝突を避けられる。効果は制約と同じで、
		// 複合 FK の参照先としても使える。CREATE ... IF NOT EXISTS 自体が冪等なので DO ブロックは要らない。
		//
		// workspaces.slug は URL に出るグローバル一意の識別子。
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_workspaces_slug ON workspaces (slug);`,
		// spaces.key はワークスペース内で一意。
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_spaces_workspace_key ON spaces (workspace_id, "key");`,
		// (workspace_id, id) の複合 UNIQUE は複合 FK の参照先として必要（id 単独の PK では
		// FK の参照列に (workspace_id, id) を指定できない）。実データ上は id の PK で一意なので、
		// 冗長に見えても「テナント越えを FK で塞ぐ」ための足場として張る。
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_spaces_workspace_id ON spaces (workspace_id, id);`,
		// pages (workspace_id, id) は blocks / page_paths からの複合 FK の参照先。
		// 親ページの FK は下の uq_pages_workspace_space_id（space_id 込み）を参照先にするが、
		// こちらは落とせない: fk_blocks_page / fk_page_paths_page / fk_page_paths_ancestor の 3 本が
		// 参照先として使っている（blocks の親を張り替えたときは旧索引に参照が無かったので落とせた。
		// pages では事情が違う）。space_id を持たないテーブルからページを参照するには
		// この (workspace_id, id) の形が要る。
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_pages_workspace_id ON pages (workspace_id, id);`,
		// 親ページの FK を「同じスペース」まで絞るための足場。
		// ページの木はスペースの中で閉じる（スペースはページの入れ物であり、木がスペースを
		// またぐとパンくず・サブツリー取得・スペース単位の権限がすべて破綻する）。
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_pages_workspace_space_id ON pages (workspace_id, space_id, id);`,
		// blocks の自己参照（親ブロック）も複合 FK にするため、blocks にも同じ足場を張る。
		// ここに page_id まで含めるのは、親ブロックを「同じワークスペース」ではなく
		// 「同じページ」に限定するため（別ページのブロックを親にできると、そのページを削除したときに
		// 親の ON DELETE CASCADE が別ページの本文まで道連れにする）。
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_blocks_workspace_page_id ON blocks (workspace_id, page_id, id);`,

		// --- 複合 FK（テナント越えの親子を作れなくする）---
		// spaces.workspace_id → workspaces.id。ワークスペースの物理削除で配下も消える
		// （運用ではアーカイブを使う想定で、物理削除は例外的な操作）。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_spaces_workspace' AND conrelid = 'spaces'::regclass) THEN
				ALTER TABLE spaces
					ADD CONSTRAINT fk_spaces_workspace
					FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE;
			END IF;
		END $$;`,
		// pages は「同じワークスペースの space」にしか属せない。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_pages_space' AND conrelid = 'pages'::regclass) THEN
				ALTER TABLE pages
					ADD CONSTRAINT fk_pages_space
					FOREIGN KEY (workspace_id, space_id) REFERENCES spaces (workspace_id, id) ON DELETE CASCADE;
			END IF;
		END $$;`,
		// pages の親は「同じワークスペースの、同じスペースの」page だけ。親の物理削除で子孫も消える。
		// ページの木はスペースの中で閉じる。スペースはページの入れ物であり、木がスペースをまたぐと
		// パンくず（祖先をたどると別スペースに出る）・サブツリー一括取得・スペース単位の権限が
		// すべて破綻するため、space_id まで一致を要求する。
		// workspace だけの一致だと、スペース A のページがスペース B のページを親に持ててしまい、
		// スペース B を消したときに fk_pages_space の CASCADE で B のページが消え、続けて
		// こちらの CASCADE がスペース A に残るはずの子ページまで道連れにする。
		// parent_id は NULL 可（ルート）。複合 FK は既定の MATCH SIMPLE なので、
		// 参照列に 1 つでも NULL があれば検査自体が行われない ＝ ルートページは素通りする。
		// これは意図どおり: ルートの workspace_id / space_id は fk_pages_space 側で必ず検査されるため、
		// テナント越え・スペース越えの抜け道にはならない。
		//
		// 副作用（意図した挙動）: ページを別スペースへ移すときは、子孫の space_id も同じ文で
		// 更新しないと FK 違反になる。木の一部だけがスペースをまたぐ「中途半端な移動」を DB が防ぐ。
		//
		// 旧定義の判定に pg_get_constraintdef の文字列一致は使えない。'workspace_id' が部分文字列として
		// 'space_id' を含むため、旧定義（workspace_id と parent_id だけ）でも LIKE '%space_id%' が真になり、
		// 張り替えが起きないまま黙って旧定義が残る。blocks を 'page_id' で判定したときは
		// この衝突が無かったので文字列一致で足りていた。ここは参照元の列そのものを見て判定する。
		`DO $$ BEGIN
			IF EXISTS (
				SELECT 1 FROM pg_constraint c
				WHERE c.conname = 'fk_pages_parent' AND c.conrelid = 'pages'::regclass
				  AND NOT EXISTS (
					SELECT 1 FROM pg_attribute a
					WHERE a.attrelid = c.conrelid AND a.attnum = ANY (c.conkey) AND a.attname = 'space_id'
				  )
			) THEN
				ALTER TABLE pages DROP CONSTRAINT fk_pages_parent; -- 旧定義（workspace のみ一致）を張り替える
			END IF;
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_pages_parent' AND conrelid = 'pages'::regclass) THEN
				ALTER TABLE pages
					ADD CONSTRAINT fk_pages_parent
					FOREIGN KEY (workspace_id, space_id, parent_id)
					REFERENCES pages (workspace_id, space_id, id) ON DELETE CASCADE;
			END IF;
		END $$;`,
		// blocks は「同じワークスペースの page」にしか属せない。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_blocks_page' AND conrelid = 'blocks'::regclass) THEN
				ALTER TABLE blocks
					ADD CONSTRAINT fk_blocks_page
					FOREIGN KEY (workspace_id, page_id) REFERENCES pages (workspace_id, id) ON DELETE CASCADE;
			END IF;
		END $$;`,
		// blocks の親は「同じワークスペースの、同じページの」block だけ。
		// ブロックの木は 1 ページの中で閉じるものなので、page_id まで一致を要求する。
		// workspace だけを一致させると、ページ A のブロックをページ B のブロックの親にでき、
		// ページ A を消したときに ON DELETE CASCADE がページ B の本文まで消してしまう。
		// MATCH SIMPLE の扱いは pages と同じで、parent_id が NULL（トップレベル）なら検査されない。
		// その場合の workspace_id / page_id の正しさは fk_blocks_page 側で担保される。
		`DO $$ BEGIN
			IF EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'fk_blocks_parent' AND conrelid = 'blocks'::regclass
				  AND pg_get_constraintdef(oid) NOT LIKE '%page_id%'
			) THEN
				ALTER TABLE blocks DROP CONSTRAINT fk_blocks_parent; -- 旧定義（workspace のみ一致）を張り替える
			END IF;
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_blocks_parent' AND conrelid = 'blocks'::regclass) THEN
				ALTER TABLE blocks
					ADD CONSTRAINT fk_blocks_parent
					FOREIGN KEY (workspace_id, page_id, parent_id)
					REFERENCES blocks (workspace_id, page_id, id) ON DELETE CASCADE;
			END IF;
		END $$;`,
		// 旧定義で作られた (workspace_id, id) の索引を落とす。FK を張り替えた「後」でないと、
		// その索引に依存している制約が残っていて DROP が失敗する（順序が意味を持つ）。
		`DROP INDEX IF EXISTS uq_blocks_workspace_id;`,
		// page_paths / page_snapshots は pages から導ける派生データ。ページが消えたら一緒に消す。
		// page_paths は 1 行で「子孫」と「祖先」の 2 ページを組にするため、単独 FK を 2 本張るだけでは
		// 別ワークスペースの 2 ページを組にした行が作れてしまう（両方の FK を通ってしまう）。
		// 行自身の workspace_id を軸にした複合 FK にして、組になる 2 ページが同じワークスペースに
		// 属することを DB 側で保証する。
		//
		// FK で守るのは「組になる 2 ページが実在し、同じワークスペースに属すること」まで。
		// 1 行だけで判定できる depth の不変条件は下の ck_page_paths_depth で別に塞ぐ。
		//
		// 一方「depth が実際の親子の距離と一致するか、祖先の連鎖に抜けや余りが無いか」は DB では守らない。
		// それは 1 行を見ても判定できず、pages の木をたどって初めて分かる複数行にまたがる不変条件で、
		// 宣言的な制約（行ごとの CHECK / FK）では表せないため。この表は pages.parent_id から導ける
		// 派生データなので、正本である pages 側の制約で木の形を守り、closure 全体の整合は行を書く側の
		// 責務とする。なお page_paths は常に FK の子側で、この表の行が壊れても他の行を CASCADE で
		// 消すことはない（壊れ方が表示の乱れに閉じ、他のデータを失わない）。
		`DO $$ BEGIN
			IF EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'fk_page_paths_page' AND conrelid = 'page_paths'::regclass
				  AND pg_get_constraintdef(oid) NOT LIKE '%workspace_id%'
			) THEN
				ALTER TABLE page_paths DROP CONSTRAINT fk_page_paths_page; -- 旧定義（単独 FK）を張り替える
			END IF;
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_page_paths_page' AND conrelid = 'page_paths'::regclass) THEN
				ALTER TABLE page_paths
					ADD CONSTRAINT fk_page_paths_page
					FOREIGN KEY (workspace_id, page_id) REFERENCES pages (workspace_id, id) ON DELETE CASCADE;
			END IF;
		END $$;`,
		`DO $$ BEGIN
			IF EXISTS (
				SELECT 1 FROM pg_constraint
				WHERE conname = 'fk_page_paths_ancestor' AND conrelid = 'page_paths'::regclass
				  AND pg_get_constraintdef(oid) NOT LIKE '%workspace_id%'
			) THEN
				ALTER TABLE page_paths DROP CONSTRAINT fk_page_paths_ancestor;
			END IF;
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_page_paths_ancestor' AND conrelid = 'page_paths'::regclass) THEN
				ALTER TABLE page_paths
					ADD CONSTRAINT fk_page_paths_ancestor
					FOREIGN KEY (workspace_id, ancestor_id) REFERENCES pages (workspace_id, id) ON DELETE CASCADE;
			END IF;
		END $$;`,
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'fk_page_snapshots_page' AND conrelid = 'page_snapshots'::regclass) THEN
				ALTER TABLE page_snapshots
					ADD CONSTRAINT fk_page_snapshots_page
					FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE;
			END IF;
		END $$;`,

		// --- CHECK ---
		// 自分自身を親にできない（1 行で閉じた循環を作らせない。多段の循環はアプリ側で検出する）。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_pages_parent_not_self' AND conrelid = 'pages'::regclass) THEN
				ALTER TABLE pages
					ADD CONSTRAINT ck_pages_parent_not_self CHECK (parent_id IS NULL OR parent_id <> id);
			END IF;
		END $$;`,
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_blocks_parent_not_self' AND conrelid = 'blocks'::regclass) THEN
				ALTER TABLE blocks
					ADD CONSTRAINT ck_blocks_parent_not_self CHECK (parent_id IS NULL OR parent_id <> id);
			END IF;
		END $$;`,
		// closure table の 1 行だけで判定できる不変条件: depth は祖先までの距離なので負にならず、
		// depth=0 の行は自分自身を指す行「だけ」（逆に自己参照の行は必ず depth=0）。
		// パンくずは ORDER BY depth で組み立てるため、ここが崩れると pages.parent_id（正本）は
		// 正しいのに表示だけが壊れ、原因を追いにくい形で顕在化する。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_page_paths_depth' AND conrelid = 'page_paths'::regclass) THEN
				ALTER TABLE page_paths
					ADD CONSTRAINT ck_page_paths_depth
					CHECK (depth >= 0 AND (depth = 0) = (page_id = ancestor_id));
			END IF;
		END $$;`,
		// URL に出る識別子は空文字禁止・長さ上限（アプリ側検証と二重の壁）。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_workspaces_slug_len' AND conrelid = 'workspaces'::regclass) THEN
				ALTER TABLE workspaces
					ADD CONSTRAINT ck_workspaces_slug_len CHECK (char_length(slug) BETWEEN 1 AND 64);
			END IF;
		END $$;`,
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_spaces_key_len' AND conrelid = 'spaces'::regclass) THEN
				ALTER TABLE spaces
					ADD CONSTRAINT ck_spaces_key_len CHECK (char_length("key") BETWEEN 1 AND 64);
			END IF;
		END $$;`,
		// position は空文字だと順序として意味を持たない（fracindex は空文字を返さない）。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_pages_position_not_empty' AND conrelid = 'pages'::regclass) THEN
				ALTER TABLE pages
					ADD CONSTRAINT ck_pages_position_not_empty CHECK ("position" <> '');
			END IF;
		END $$;`,
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_blocks_position_not_empty' AND conrelid = 'blocks'::regclass) THEN
				ALTER TABLE blocks
					ADD CONSTRAINT ck_blocks_position_not_empty CHECK ("position" <> '');
			END IF;
		END $$;`,
		// page_snapshots.doc は tiptap のドキュメント JSON（object かつ type='doc'）に限る。
		// 壊れた snapshot は読み取りキャッシュとしてそのまま返り、エディタがページを開けなくなるため、
		// rich_documents.doc と同じ形で入口を塞ぐ。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_page_snapshots_doc' AND conrelid = 'page_snapshots'::regclass) THEN
				ALTER TABLE page_snapshots
					ADD CONSTRAINT ck_page_snapshots_doc
					CHECK (jsonb_typeof(doc) = 'object' AND doc->>'type' = 'doc');
			END IF;
		END $$;`,
		// attrs は ProseMirror の attrs なので必ず object（属性が無いノードでも {}）。
		// inline は葉ノードの content 配列。容器ノードでは NULL にする。
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_blocks_attrs_object' AND conrelid = 'blocks'::regclass) THEN
				ALTER TABLE blocks
					ADD CONSTRAINT ck_blocks_attrs_object CHECK (jsonb_typeof(attrs) = 'object');
			END IF;
		END $$;`,
		`DO $$ BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'ck_blocks_inline_array' AND conrelid = 'blocks'::regclass) THEN
				ALTER TABLE blocks
					ADD CONSTRAINT ck_blocks_inline_array CHECK (inline IS NULL OR jsonb_typeof(inline) = 'array');
			END IF;
		END $$;`,

		// --- 並び順の一意性（部分 UNIQUE は制約にできないので UNIQUE INDEX で張る。
		// CREATE ... IF NOT EXISTS 自体が冪等なので DO ブロックでは包まない）---
		// 同じ親の中で position が重複しないこと。ページはアーカイブ済みを除外する
		// （アーカイブは「一覧から隠す」だけで行は残るため、現役の並びだけを守る）。
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_pages_parent_position
			ON pages (parent_id, "position") WHERE archived_at IS NULL;`,
		// ルート直下（parent_id IS NULL）は上の索引では守れない。UNIQUE 索引は NULL 同士を
		// 別物として扱うため、parent_id が NULL の行同士は何度でも同じ position を持ててしまう。
		// ルートの並びはスペース単位なので、スペースを軸にした部分 UNIQUE を別に張る。
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_pages_space_position
			ON pages (space_id, "position") WHERE parent_id IS NULL AND archived_at IS NULL;`,
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_blocks_parent_position
			ON blocks (parent_id, "position");`,
		// ブロックも同じ理由で、ページ直下（parent_id IS NULL）はページを軸に守る。
		`CREATE UNIQUE INDEX IF NOT EXISTS uq_blocks_page_position
			ON blocks (page_id, "position") WHERE parent_id IS NULL;`,
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
