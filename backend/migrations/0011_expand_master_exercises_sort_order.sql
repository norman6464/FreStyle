-- 0011 (Expand): master_exercises に sort_order を追加し order_index から backfill。order_index は残す。
--
-- Expand-Contract の Expand フェーズ (FRESTYLE-183 / 親 FRESTYLE-181)。並び順カラムを sort_order に
-- 統一する。additive なので、稼働中の backend(order_index を Order by で読むだけ)には無影響。
-- master_exercises の唯一の writer は教材リポ seed.py。Contract で seed.py を sort_order に切替え、
-- backend を sort_order 読みにデプロイし切ってから 0012 で order_index を削除する。
--
-- 冪等: 列が無ければ追加し、backfill は sort_order が既定(0)で order_index が非0の行だけ更新する。
--
-- 適用: frestyle-infrastructure リポで
--   make apply-migration-supabase FILE=../FreStyle/backend/migrations/0011_expand_master_exercises_sort_order.sql \
--        DATABASE_URL_SECRET_NAME=frestyle-prod/database-url

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'master_exercises' AND column_name = 'sort_order'
    ) THEN
        ALTER TABLE master_exercises ADD COLUMN sort_order integer NOT NULL DEFAULT 0;
    END IF;
END $$;

UPDATE master_exercises
SET sort_order = order_index
WHERE sort_order = 0 AND order_index <> 0;
