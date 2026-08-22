-- 0021 (Contract): users の旧カラム cognito_sub / role / onboarded_at を削除する。
--
-- Expand-Contract の Contract フェーズ（FRESTYLE-311）。正規化（roles マスタ + role_id FK /
-- user_oidc_identities）を導入し、起動時バックフィルで既存データを新テーブルへ移送済み。
-- 読み替えフェーズ（PR #2248）に続き、旧カラムへの書き込みも止めた Contract 版コードを
-- 全 ECS タスクにデプロイし切ったので、旧カラムを物理削除して正規化を完了する。
--
-- ⚠️ 適用順序が重要: 必ず Contract 版コード（旧カラムを読み書きしない版・本 PR）を全タスクに
--    デプロイし切ってから適用すること。先に DROP すると、まだ旧カラムを読む/書く旧タスクが壊れる。
--    起動時バックフィル BackfillUserNormalization は旧カラム依存部をカラム存在チェックでガード
--    しているため、本 migration 適用後の起動でも安全に no-op になる。
--
-- 冪等: 各カラムが存在するときだけ削除する（DROP COLUMN は付随する index / 制約も自動削除する）。
-- 直列化: 起動時マイグレーション（internal/infra/database/migrate.go の Migrate）と同じ
--   pg_advisory_xact_lock(4915311) を取得してから DROP する。これにより、起動タスクの
--   BackfillUserNormalization（cognito_sub 列がある間だけ当該列を読む）と本 DROP が同時実行されて
--   「読んでいる最中に列が消える」競合を防ぐ。ロックは DO ブロックのトランザクション終了で自動解放。
-- スキーマ限定: information_schema / ALTER TABLE を public スキーマの users に限定する。
--
-- 適用: frestyle-infrastructure リポで
--   make apply-migration-supabase FILE=../FreStyle/backend/migrations/0021_contract_drop_users_legacy_columns.sql \
--        DATABASE_URL_SECRET_NAME=frestyle-prod/database-url

DO $$
BEGIN
    PERFORM pg_advisory_xact_lock(4915311);

    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'cognito_sub'
    ) THEN
        ALTER TABLE public.users DROP COLUMN cognito_sub;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'role'
    ) THEN
        ALTER TABLE public.users DROP COLUMN role;
    END IF;

    -- onboarded_at はドメインモデルから既に撤去済み（読み書きするコードは無い）。
    -- 物理カラムが過去スキーマの名残として残っている環境のための掃除。
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'onboarded_at'
    ) THEN
        ALTER TABLE public.users DROP COLUMN onboarded_at;
    END IF;
END $$;
