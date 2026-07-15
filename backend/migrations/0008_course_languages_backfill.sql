-- 0008: 既存コースの language を backfill する（FRESTYLE-114）。
--
-- courses.language 列自体は GORM AutoMigrate（domain.Course の Language 追加）が
-- デプロイ時に自動作成する。この SQL は本番の既存 22 コースへ初期値を入れるだけの
-- 冪等な UPDATE（何度流しても同じ結果）。言語が主題でないコース
-- （FreStyle プロダクト入門 / ステージング検証 / Design Doc 入門）は空のまま = バッジ非表示。
--
-- 適用: infra リポで
--   make apply-migration-supabase FILE=.../0008_course_languages_backfill.sql \
--        DATABASE_URL_SECRET_NAME=frestyle-prod/database-url
--
-- 列が未作成の環境（backend デプロイ前）に誤って流した場合に備え、
-- 列が存在するときだけ実行する。

DO $$
BEGIN
  IF EXISTS (
    SELECT 1 FROM information_schema.columns
    WHERE table_name = 'courses' AND column_name = 'language'
  ) THEN
    UPDATE courses SET language = 'web'        WHERE id = 1  AND title LIKE 'Web 基礎%';
    UPDATE courses SET language = 'git'        WHERE id = 2  AND title LIKE 'Git%';
    UPDATE courses SET language = 'docker'     WHERE id = 3  AND title LIKE 'Docker%';
    UPDATE courses SET language = 'linux'      WHERE id = 4  AND title LIKE 'Linux%';
    UPDATE courses SET language = 'go'         WHERE id = 5  AND title LIKE 'FreStyle バックエンド入門%';
    UPDATE courses SET language = 'go'         WHERE id = 6  AND title LIKE 'Go 言語%';
    UPDATE courses SET language = 'go'         WHERE id = 7  AND title LIKE 'クリーンアーキテクチャ%';
    UPDATE courses SET language = 'go'         WHERE id = 8  AND title LIKE 'ヘキサゴナル%';
    UPDATE courses SET language = 'go'         WHERE id = 9  AND title LIKE 'レイヤード%';
    UPDATE courses SET language = 'postgresql' WHERE id = 10 AND title LIKE 'PostgreSQL%';
    UPDATE courses SET language = 'openapi'    WHERE id = 12 AND title LIKE 'OpenAPI%';
    UPDATE courses SET language = 'go'         WHERE id = 15 AND title LIKE 'テスト徹底入門%';
    UPDATE courses SET language = 'go'         WHERE id = 16 AND title LIKE 'Go API 設計%';
    UPDATE courses SET language = 'terraform'  WHERE id = 17 AND title LIKE 'Terraform%';
    UPDATE courses SET language = 'go'         WHERE id = 18 AND title LIKE 'FreStyle バックエンド コードリーディング%';
    UPDATE courses SET language = 'aws'        WHERE id = 19 AND title LIKE 'AWS%';
    UPDATE courses SET language = 'web'        WHERE id = 20 AND title LIKE 'Web セキュリティ%';
    UPDATE courses SET language = 'go'         WHERE id = 21 AND title LIKE 'クリーンコード%';
    UPDATE courses SET language = 'go'         WHERE id = 22 AND title LIKE 'FreStyle 自作リンター%';
  END IF;
END $$;
