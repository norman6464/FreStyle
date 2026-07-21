-- 0014 (Contract): 互換 VIEW teaching_materials を削除する。
--
-- FRESTYLE-184（親 FRESTYLE-181）。0013 でテーブルを course_chapters に改名した際、
-- ローリングデプロイ中の旧タスクを壊さないために旧名の互換 VIEW を作成した。
--
-- ⚠️ 適用条件: 以下がすべて完了してから適用すること。
--   (a) backend が course_chapters / sort_order を参照するコードでデプロイ済み（全タスク入替済）
--   (b) 教材リポ seed-courses.py が course_chapters / sort_order に切替済
-- 先に DROP すると、まだ旧名を参照するタスク・seed が壊れる。
--
-- 冪等: VIEW が存在するときだけ削除する（IF EXISTS）。
--
-- 適用: frestyle-infrastructure リポで
--   make apply-migration-supabase FILE=../FreStyle/backend/migrations/0014_drop_teaching_materials_compat_view.sql \
--        DATABASE_URL_SECRET_NAME=frestyle-prod/database-url

DROP VIEW IF EXISTS teaching_materials;
