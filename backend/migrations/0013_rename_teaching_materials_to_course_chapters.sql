-- 0013: teaching_materials → course_chapters / order_in_course → sort_order（互換 VIEW 付き）
--
-- FRESTYLE-184（親 FRESTYLE-181）。章を表すテーブルの語彙を chapter に統一する。
-- 命名規約: 子エンティティ(1:N) は親を接頭辞にして所有と 1:N を名前で示す → course_chapters
-- （course 1 : N chapter）。並び順は courses.sort_order と揃えて sort_order に統一。
--
-- 【無停止のための互換 VIEW】
-- テーブル改名は後方互換でなく、ECS ローリングデプロイ中は新旧タスクが同時に稼働する。そこで
-- 改名後に旧名 teaching_materials を VIEW として作り直し、旧コードの SELECT/INSERT/UPDATE/DELETE
-- をそのまま通す（PostgreSQL の自動更新可能ビュー: 単一テーブル・単純列参照・列別名は可）。
-- 新コードをデプロイし切ったあと 0014 でこの VIEW を DROP する。
--
-- 冪等: 実テーブルが旧名のときだけ改名し、列も未改名のときだけ改名する。VIEW は CREATE OR REPLACE。
--
-- 適用: frestyle-infrastructure リポで
--   make apply-migration-supabase FILE=../FreStyle/backend/migrations/0013_rename_teaching_materials_to_course_chapters.sql \
--        DATABASE_URL_SECRET_NAME=frestyle-prod/database-url

DO $$
BEGIN
    -- 1) テーブル改名（実テーブルが旧名で、新名がまだ無いときだけ）
    IF EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'teaching_materials' AND table_type = 'BASE TABLE'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'course_chapters'
    ) THEN
        ALTER TABLE teaching_materials RENAME TO course_chapters;
    END IF;

    -- 2) 並び順カラム改名（未改名のときだけ）
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'course_chapters' AND column_name = 'order_in_course'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'course_chapters' AND column_name = 'sort_order'
    ) THEN
        ALTER TABLE course_chapters RENAME COLUMN order_in_course TO sort_order;
    END IF;
END $$;

-- 3) 旧名の互換 VIEW（デプロイ完了後に 0014 で DROP する）
CREATE OR REPLACE VIEW teaching_materials AS
SELECT id,
       company_id,
       created_by_user_id,
       title,
       content,
       is_published,
       created_at,
       updated_at,
       course_id,
       sort_order AS order_in_course
FROM course_chapters;
