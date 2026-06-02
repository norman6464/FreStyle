-- =====================================================================
-- 0002 — Unify code exercises (php_exercises → master_exercises)
-- =====================================================================
-- 目的:
--   PR-H3 で導入した「言語非依存の master_exercises」スキーマに既存の
--   php_exercises テーブルをリネーム + 列拡張で移行する。新規テーブル
--   (company_exercises / exercise_submissions) は GORM AutoMigrate が
--   起動時に作るため、本マイグレーションは既存テーブルの ALTER のみ。
--
-- 想定 DB: AWS RDS PostgreSQL (frestyle-prod-rds-postgres)
-- 実行手順: EC2 踏み台 (test) 経由で psql 接続 → 本ファイルを実行。
--   詳細は frestyle-pdm/docs/migration/0004-unify-exercises.md を参照。
--
-- 冪等性:
--   - すべて IF NOT EXISTS / IF EXISTS で複数回実行しても安全
--   - 既に master_exercises にリネーム済の環境では BEGIN..COMMIT 内の
--     RENAME ブランチを skip する（DO $$ ... $$ で存在チェック）
--
-- ロールバック:
--   ALTER TABLE master_exercises RENAME TO php_exercises;
--   各 ADD COLUMN を DROP COLUMN で戻す（データ損失なし）
--
-- 関連:
--   - 実装側の変更は norman6464/FreStyle PR-H3 で merged 予定
--   - AutoMigrate は新規テーブル作成のみ、ALTER は本 SQL で先行適用
-- =====================================================================

BEGIN;

-- 1. 既存 php_exercises に新カラム追加（冪等）
ALTER TABLE IF EXISTS php_exercises
    ADD COLUMN IF NOT EXISTS language     VARCHAR(32) NOT NULL DEFAULT 'php',
    ADD COLUMN IF NOT EXISTS slug         VARCHAR(64),
    ADD COLUMN IF NOT EXISTS difficulty   SMALLINT    NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS is_published BOOLEAN     NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS chapter_id   BIGINT      NULL;

-- 2. slug を id ベースで backfill（NULL の行のみ更新）
UPDATE php_exercises
   SET slug = 'php-' || id
 WHERE slug IS NULL;

-- 3. slug を NOT NULL + UNIQUE に昇格
ALTER TABLE IF EXISTS php_exercises
    ALTER COLUMN slug SET NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uniq_php_exercises_slug ON php_exercises(slug);
CREATE INDEX        IF NOT EXISTS idx_php_exercises_language ON php_exercises(language);

-- 4. テーブル / インデックス / シーケンスを master_exercises 系にリネーム
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'php_exercises') THEN
        ALTER TABLE php_exercises RENAME TO master_exercises;
    END IF;

    IF EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'uniq_php_exercises_slug') THEN
        ALTER INDEX uniq_php_exercises_slug RENAME TO uniq_master_exercises_slug;
    END IF;

    IF EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_php_exercises_language') THEN
        ALTER INDEX idx_php_exercises_language RENAME TO idx_master_exercises_language;
    END IF;

    IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'php_exercises_id_seq' AND relkind = 'S') THEN
        ALTER SEQUENCE php_exercises_id_seq RENAME TO master_exercises_id_seq;
    END IF;
END $$;

COMMIT;

-- 動作確認用クエリ（実行後に手動で叩いて検証）:
-- SELECT id, slug, language, title FROM master_exercises ORDER BY order_index LIMIT 5;
-- SELECT COUNT(*) FROM master_exercises;
-- \d master_exercises
