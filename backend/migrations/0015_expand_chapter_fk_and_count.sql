-- 0015 (Expand): 章 FK / カウンタ列を chapter 語彙へ移行するための追加 + backfill。旧列は残す。
--
-- FRESTYLE-185（親 FRESTYLE-181）。Phase 4a で参照先が course_chapters（実体 = chapter）に
-- なったため、FK を chapter_id に、日次カウンタを chapter_count に統一する。
-- additive なので稼働中の旧 backend（旧列を参照）には無影響。
--
-- 新列側の UNIQUE index も同時に作る。Contract で upsert の ON CONFLICT を新列へ
-- 切り替えるには、先に新列側の一意制約が存在している必要があるため。
--
-- 冪等: 列・index は存在チェック付き。backfill は未設定行のみ更新する。
--
-- 適用: frestyle-infrastructure リポで
--   make apply-migration-supabase FILE=../FreStyle/backend/migrations/0015_expand_chapter_fk_and_count.sql \
--        DATABASE_URL_SECRET_NAME=frestyle-prod/database-url

-- 1) user_chapter_views.teaching_material_id → chapter_id
ALTER TABLE user_chapter_views ADD COLUMN IF NOT EXISTS chapter_id bigint;
UPDATE user_chapter_views SET chapter_id = teaching_material_id WHERE chapter_id IS NULL;
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM user_chapter_views WHERE chapter_id IS NULL) THEN
        ALTER TABLE user_chapter_views ALTER COLUMN chapter_id SET NOT NULL;
    END IF;
END $$;
CREATE UNIQUE INDEX IF NOT EXISTS ux_user_chapter_views_user_chapter
    ON user_chapter_views (user_id, chapter_id);

-- 2) user_lesson_progress.teaching_material_id → chapter_id
ALTER TABLE user_lesson_progress ADD COLUMN IF NOT EXISTS chapter_id bigint;
UPDATE user_lesson_progress SET chapter_id = teaching_material_id WHERE chapter_id IS NULL;
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM user_lesson_progress WHERE chapter_id IS NULL) THEN
        ALTER TABLE user_lesson_progress ALTER COLUMN chapter_id SET NOT NULL;
    END IF;
END $$;
CREATE UNIQUE INDEX IF NOT EXISTS ux_user_chapter_progress_user_chapter
    ON user_lesson_progress (user_id, chapter_id);

-- 3) user_daily_activities.lesson_count → chapter_count
ALTER TABLE user_daily_activities ADD COLUMN IF NOT EXISTS chapter_count integer NOT NULL DEFAULT 0;
UPDATE user_daily_activities SET chapter_count = lesson_count WHERE chapter_count = 0 AND lesson_count <> 0;
