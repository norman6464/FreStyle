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
--   本マイグレーションが挿入した行 = 「OrderIndex=1 / InputText='' で既に master_exercises に
--   存在していた expected_output と一致するもの」のみを取り消す。 全件 DELETE は運営や PR-W で
--   後から追加された手動 example を巻き込むため絶対に避けること。
--
--   DELETE FROM master_exercise_examples mee
--    WHERE mee.order_index = 1
--      AND mee.input_text  = ''
--      AND EXISTS (
--          SELECT 1
--            FROM master_exercises me
--           WHERE me.id = mee.exercise_id
--             AND COALESCE(me.expected_output, '') = mee.expected_output
--      );
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
