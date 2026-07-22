-- 0016 (Contract): 章 FK / カウンタの旧列を削除する。
--
-- FRESTYLE-185（親 FRESTYLE-181）。0015 で chapter_id / chapter_count を追加し dual-write して
-- きた。本 migration で旧列を削除し chapter 語彙へ一本化する。
--
-- ⚠️ 適用条件: 新列のみを読み書きする Contract 版コードを全タスクにデプロイし切ってから適用する。
--    先に DROP すると、まだ旧列を参照する Expand 版タスクが壊れる。
--
-- user_chapter_views は (user_id, teaching_material_id) が複合 PK のため、旧列を落とす前に
-- PK を張り替える。0015 で作成済みの UNIQUE index を `USING INDEX` でそのまま PK に昇格させ、
-- 余分な index を作らない。
--
-- 冪等: 旧列が存在するときだけ実行する。
--
-- 適用: frestyle-infrastructure リポで
--   make apply-migration-supabase FILE=../FreStyle/backend/migrations/0016_contract_drop_old_chapter_fk_and_count.sql \
--        DATABASE_URL_SECRET_NAME=frestyle-prod/database-url

-- 1) user_chapter_views: PK を (user_id, chapter_id) に張り替えてから旧列を削除
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'user_chapter_views' AND column_name = 'teaching_material_id'
    ) THEN
        ALTER TABLE user_chapter_views DROP CONSTRAINT IF EXISTS user_chapter_views_pkey;
        ALTER TABLE user_chapter_views
            ADD CONSTRAINT user_chapter_views_pkey PRIMARY KEY USING INDEX ux_user_chapter_views_user_chapter;
        ALTER TABLE user_chapter_views DROP COLUMN teaching_material_id;
    END IF;
END $$;

-- 2) user_lesson_progress: 旧 UNIQUE index を落として旧列を削除（PK は id のまま）
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'user_lesson_progress' AND column_name = 'teaching_material_id'
    ) THEN
        DROP INDEX IF EXISTS ux_user_lesson;
        ALTER TABLE user_lesson_progress DROP COLUMN teaching_material_id;
    END IF;
END $$;

-- 3) user_daily_activities: 旧カウンタ列を削除
ALTER TABLE user_daily_activities DROP COLUMN IF EXISTS lesson_count;
