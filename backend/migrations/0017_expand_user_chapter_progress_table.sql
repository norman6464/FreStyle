-- 0017 (Expand): user_lesson_progress の後継テーブル user_chapter_progress を作成し全行コピーする。
--
-- FRESTYLE-186（親 FRESTYLE-181）。DB 上の「章」語彙で唯一残る "lesson" を chapter に統一する。
--
-- 【なぜテーブル改名(+互換 VIEW)ではないのか】
-- このテーブルへの書き込みは GORM の upsert（INSERT ... ON CONFLICT (user_id, chapter_id)
-- DO NOTHING）。PostgreSQL は VIEW に対する ON CONFLICT をサポートしないため、4a で使った
-- 「改名 + 旧名の互換 VIEW」方式が使えない（INSTEAD OF トリガを付けても ON CONFLICT は不可）。
-- そこで新テーブルを作り、移行期は 2 テーブルへ dual-write する Expand-Contract とする。
--
-- 冪等: テーブル・index は IF NOT EXISTS、コピーは ON CONFLICT DO NOTHING。
--
-- 適用: frestyle-infrastructure リポで
--   make apply-migration-supabase FILE=../FreStyle/backend/migrations/0017_expand_user_chapter_progress_table.sql \
--        DATABASE_URL_SECRET_NAME=frestyle-prod/database-url

CREATE TABLE IF NOT EXISTS user_chapter_progress (
    id           bigserial PRIMARY KEY,
    user_id      bigint      NOT NULL,
    chapter_id   bigint      NOT NULL,
    course_id    bigint      NOT NULL,
    completed_at timestamptz NOT NULL,
    created_at   timestamptz
);

CREATE UNIQUE INDEX IF NOT EXISTS ux_user_chapter_progress
    ON user_chapter_progress (user_id, chapter_id);
CREATE INDEX IF NOT EXISTS idx_user_chapter_progress_course_id
    ON user_chapter_progress (course_id);

-- 既存の完了記録をコピー（id を保持して再適用時も重複しないようにする）
INSERT INTO user_chapter_progress (id, user_id, chapter_id, course_id, completed_at, created_at)
SELECT id, user_id, chapter_id, course_id, completed_at, created_at
FROM user_lesson_progress
WHERE chapter_id IS NOT NULL
ON CONFLICT DO NOTHING;

-- id シーケンスをコピー済みの最大値に合わせる（以後の INSERT が衝突しないように）
SELECT setval(
    pg_get_serial_sequence('user_chapter_progress', 'id'),
    GREATEST(COALESCE((SELECT MAX(id) FROM user_chapter_progress), 1), 1)
);
