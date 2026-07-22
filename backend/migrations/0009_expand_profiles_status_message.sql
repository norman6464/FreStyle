-- 0009 (Expand): profiles に status_message を追加し status から backfill。status は残す。
--
-- Expand-Contract の Expand フェーズ (FRESTYLE-182 / 親 FRESTYLE-181)。列リネームは後方互換で
-- ないため、まず新列を追加して両立させる。additive なので、稼働中の旧 backend(status を明示列で
-- SELECT / 更新している)には無影響。次にコードを dual-write(status と status_message の両方に書く)
-- にデプロイし、最後の Contract フェーズ(0010)で status を削除する。
--
-- 冪等: 列が無ければ追加し、backfill は status_message が空で status が非空の行だけ更新する。
-- 何度流しても同じ状態に収束する。
--
-- 適用: frestyle-infrastructure リポで
--   make apply-migration-supabase FILE=../FreStyle/backend/migrations/0009_expand_profiles_status_message.sql \
--        DATABASE_URL_SECRET_NAME=frestyle-prod/database-url

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'profiles' AND column_name = 'status_message'
    ) THEN
        ALTER TABLE profiles ADD COLUMN status_message text NOT NULL DEFAULT '';
    END IF;
END $$;

UPDATE profiles
SET status_message = status
WHERE status_message = '' AND status <> '';
