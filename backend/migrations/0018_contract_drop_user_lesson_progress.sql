-- 0018 (Contract): 旧テーブル user_lesson_progress を削除する。
--
-- FRESTYLE-186（親 FRESTYLE-181）。0017 で後継 user_chapter_progress を作成し全行コピー、
-- 以降は 2 テーブルへ dual-write してきた。Contract 版コード（新テーブルのみ読み書き）を
-- 全タスクにデプロイし切ったので、旧テーブルを削除して「章」語彙の統一を完了する。
--
-- ⚠️ 適用条件: Contract 版コードのデプロイ完了後に適用すること。先に DROP すると、まだ
--    旧テーブルを読む dual-write 版タスクが壊れる。
--
-- 冪等: IF EXISTS。
--
-- 適用: frestyle-infrastructure リポで
--   make apply-migration-supabase FILE=../FreStyle/backend/migrations/0018_contract_drop_user_lesson_progress.sql \
--        DATABASE_URL_SECRET_NAME=frestyle-prod/database-url

DROP TABLE IF EXISTS user_lesson_progress;
