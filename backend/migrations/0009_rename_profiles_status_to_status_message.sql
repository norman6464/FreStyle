-- 0009: profiles.status を status_message に改名 (FRESTYLE-182 / 親 FRESTYLE-181)
--
-- profiles.status は「一言ステータス(自由文)」を表す列だが、状態機械(決まった状態を遷移する
-- 列挙)である invitations.status / company_applications.status / learning_reports.status と
-- 同じ名前で役割を誤解しやすい。命名規約(FRESTYLE-181)に沿い status_message に改名する。
-- GORM AutoMigrate は列リネームを追従しない(新列を足すだけ)ため本 SQL で明示的に改名する。
-- 冪等: 旧列があり新列が無いときだけ実行。何度流しても同じ状態に収束する。対象が無ければ no-op。
--
-- 適用: frestyle-infrastructure リポで
--   make apply-migration-supabase FILE=../FreStyle/backend/migrations/0009_rename_profiles_status_to_status_message.sql \
--        DATABASE_URL_SECRET_NAME=frestyle-prod/database-url

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'profiles' AND column_name = 'status'
    ) AND NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'profiles' AND column_name = 'status_message'
    ) THEN
        ALTER TABLE profiles RENAME COLUMN status TO status_message;
    END IF;
END $$;
