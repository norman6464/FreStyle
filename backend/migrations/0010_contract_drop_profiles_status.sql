-- 0010 (Contract): profiles.status を削除し status_message へ一本化する。
--
-- Expand-Contract の Contract フェーズ (FRESTYLE-182 / 親 FRESTYLE-181)。
--
-- ⚠️ 適用順序が重要: 必ず Contract 版コード(status を一切参照しない)を全タスクにデプロイし
--    切ってから本 migration を適用する。先に DROP すると、まだ status を SELECT/更新している
--    Expand 版タスクが存在する間、そのクエリが壊れる。expand(0009)とは逆順。
--
-- 冪等: status 列が存在するときだけ削除する。何度流しても同じ状態に収束する。
--
-- 適用: frestyle-infrastructure リポで
--   make apply-migration-supabase FILE=../FreStyle/backend/migrations/0010_contract_drop_profiles_status.sql \
--        DATABASE_URL_SECRET_NAME=frestyle-prod/database-url

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'profiles' AND column_name = 'status'
    ) THEN
        ALTER TABLE profiles DROP COLUMN status;
    END IF;
END $$;
