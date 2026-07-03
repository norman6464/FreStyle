-- 0007: 既存コースへの学習領域カテゴリ割当 (FRESTYLE-67)
--
-- courses.category 列は GORM AutoMigrate が追加済み(既定 '')。本 SQL は既存 22 コースに
-- 定義済みカテゴリ(domain.ValidCourseCategories / docs/course-categories.md)を割り当てる。
-- 冪等: 同じ値への UPDATE は何度流しても同じ状態に収束する。対象 id が無ければ no-op。
--
-- 適用: frestyle-infrastructure リポで
--   make apply-migration-supabase FILE=../FreStyle/backend/migrations/0007_course_categories_backfill.sql \
--        DATABASE_URL_SECRET_NAME=frestyle-prod/database-url

BEGIN;

-- 開発基礎(黄): Web 基礎 / Git / Docker / Linux
UPDATE courses SET category = 'dev-basics'   WHERE id IN (1, 2, 3, 4);

-- バックエンド開発(青): FreStyle バックエンド入門 / Go / テスト / Go API 設計 / コードリーディング / 自作リンター
UPDATE courses SET category = 'backend'      WHERE id IN (5, 6, 15, 16, 18, 22);

-- 設計・アーキテクチャ(紫): クリーンアーキ / ヘキサゴナル / レイヤード / Design Doc / クリーンコード
UPDATE courses SET category = 'architecture' WHERE id IN (7, 8, 9, 14, 21);

-- データベース(緑): PostgreSQL
UPDATE courses SET category = 'database'     WHERE id = 10;

-- プロダクト・仕様(水色): FreStyle プロダクト / OpenAPI
UPDATE courses SET category = 'product'      WHERE id IN (11, 12);

-- インフラ・クラウド(橙): ステージング検証 / Terraform / AWS
UPDATE courses SET category = 'infra'        WHERE id IN (13, 17, 19);

-- セキュリティ(赤): Web セキュリティ
UPDATE courses SET category = 'security'     WHERE id = 20;

COMMIT;
