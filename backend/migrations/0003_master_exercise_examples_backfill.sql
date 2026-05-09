-- =====================================================================
-- 0003 — master_exercise_examples テーブルへの初期 backfill
-- =====================================================================
-- 目的:
--   PR-V で導入した master_exercise_examples テーブル（GORM AutoMigrate で
--   起動時に自動生成）に対して、既存 master_exercises.expected_output を
--   「OrderIndex = 1, InputText = ''」の 1 件として backfill する。
--
--   採点ロジック (PR-W 予定) はこのテーブルの全件をテストケースとして実行する
--   ため、既存問題が空のままだと提出 / 採点が走らない。
--
--   起動時の seed (seedMasterExerciseExamples) でも同等の挿入を試みるが、
--   テーブル作成直後のレースを避けるため運用側で本 SQL を 1 度走らせる想定。
--
-- 想定 DB: AWS RDS PostgreSQL (frestyle-prod-rds-postgres)
-- 実行手順: ECS one-off (Fargate) または踏み台 EC2 経由で psql 接続して実行。
--
-- 冪等性:
--   - INSERT は WHERE NOT EXISTS で既存 example のある exercise_id を skip
--   - 何度実行しても同じ結果になる
--
-- ロールバック:
--   DELETE FROM master_exercise_examples;  -- 全件削除（テーブルは GORM 管理のため残す）
-- =====================================================================

BEGIN;

INSERT INTO master_exercise_examples
    (exercise_id, order_index, input_text, expected_output, created_at, updated_at)
SELECT
    me.id,
    1,
    '',
    COALESCE(me.expected_output, ''),
    now(),
    now()
FROM master_exercises me
WHERE NOT EXISTS (
    SELECT 1
      FROM master_exercise_examples mee
     WHERE mee.exercise_id = me.id
);

COMMIT;

-- 動作確認用クエリ:
-- SELECT me.id, me.slug, COUNT(mee.id) AS example_count
--   FROM master_exercises me
--   LEFT JOIN master_exercise_examples mee ON mee.exercise_id = me.id
--  GROUP BY me.id, me.slug
--  ORDER BY me.id;
