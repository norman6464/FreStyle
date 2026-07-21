-- 0012 (Contract): master_exercises.order_index を削除。sort_order へ一本化する。
--
-- Expand-Contract の Contract フェーズ (FRESTYLE-183 / 親 FRESTYLE-181)。
--
-- ⚠️ 適用順序が重要: 必ず (a) backend を sort_order 読みにデプロイし切り、(b) 教材リポ seed.py を
--    sort_order に切り替えた後に本 migration を適用する。先に DROP すると、まだ order_index を読む
--    旧 backend タスクや、order_index を書く旧 seed.py が壊れる。expand(0011)とは逆順。
--
-- 冪等: order_index 列が存在するときだけ削除する。
--
-- 適用: frestyle-infrastructure リポで
--   make apply-migration-supabase FILE=../FreStyle/backend/migrations/0012_contract_drop_master_exercises_order_index.sql \
--        DATABASE_URL_SECRET_NAME=frestyle-prod/database-url

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'master_exercises' AND column_name = 'order_index'
    ) THEN
        ALTER TABLE master_exercises DROP COLUMN order_index;
    END IF;
END $$;
