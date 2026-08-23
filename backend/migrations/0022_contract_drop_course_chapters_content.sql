-- 0022 (Contract): course_chapters の旧本文カラム content を削除する。
--
-- Expand-Contract の Contract フェーズ（FRESTYLE-348）。章の本文をリッチテキスト
-- （tiptap の ProseMirror JSON）の doc(jsonb) へ移行し、本番全章の doc 投入が完了、
-- フロントエンドも content を読み書きしなくなった（マージ・デプロイ済み）。
-- content(text・raw Markdown) を読み書きしない Contract 版コード（本 PR）の
-- デプロイに続き、旧カラムを物理削除して移行を完了する。
--
-- ⚠️ 適用順序が重要: 必ず Contract 版コード（content を読み書きしない版・本 PR）を全 ECS タスクに
--    デプロイし切ってから適用すること。先に DROP すると、まだ content を UPDATE する旧タスクが
--    SQLSTATE 42703（undefined_column）で壊れる。また旧 domain の gorm タグが残ったタスクが
--    再起動すると、起動時 AutoMigrate が content 列を再作成してしまう。
-- 過去 migration 0019 / 0020 は content 列を参照する一回きりのデータ移行（適用済み）で、再実行しない限り本 DROP と競合しない。
--
-- 冪等: カラムが存在するときだけ削除する（DROP COLUMN は付随する index / 制約も自動削除する）。
-- 直列化: 起動時マイグレーション（internal/infra/database/migrate.go の Migrate）と同じ
--   pg_advisory_xact_lock(4915311) を取得してから DROP する。これにより、起動タスクの
--   AutoMigrate と本 DROP が同時実行されて中途半端なスキーマになる競合を防ぐ。
--   ロックは DO ブロックのトランザクション終了で自動解放。
-- スキーマ限定: information_schema / ALTER TABLE を public スキーマの course_chapters に限定する。
--
-- 適用: frestyle-infrastructure リポで
--   make apply-migration-supabase FILE=../FreStyle/backend/migrations/0022_contract_drop_course_chapters_content.sql \
--        DATABASE_URL_SECRET_NAME=frestyle-prod/database-url

DO $$
BEGIN
    PERFORM pg_advisory_xact_lock(4915311);

    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'course_chapters' AND column_name = 'content'
    ) THEN
        ALTER TABLE public.course_chapters DROP COLUMN content;
    END IF;
END $$;
