-- ローカル開発 / 性能検証用のダミーデータ投入。
--
-- 使い方(Makefile 経由が楽):
--   make local-seed                 # small (既定)
--   make local-seed SIZE=medium
--   make local-seed SIZE=large
--
-- 直接流す場合:
--   psql "$DSN" -v size=medium -f scripts/seed-local.sql
--
-- 前提: backend を一度起動してスキーマ適用済みであること(このスクリプトは
-- テーブルを作らない。スキーマの正本は infra/database/schema/*.sql + migrations/*.sql)。
--
-- 設計方針:
--   - ORM のループ INSERT ではなく generate_series で一括生成する(桁違いに速い)
--   - setseed() で乱数を固定し、誰が何度流しても同じデータになるようにする
--     (実行計画の比較には再現性が必須。ばらつくと前後比較が意味を失う)
--   - 冪等にする。再実行時は自分が入れた範囲だけを消してから入れ直す
--   - 本物の教材本文は非公開リポが正本のため、ここでは触れずダミー文言を使う

\set ON_ERROR_STOP on

-- サイズ未指定なら small。
\if :{?size}
\else
  \set size 'small'
\endif

-- 乱数の固定。以降の random() はこの seed から決まる。
SELECT setseed(0.42);

-- 規模の定義。ダミーデータの範囲を判別できるよう、ID は SEED_ID_BASE 以降に採番する
-- (既存の実データや教材 seed と衝突させない。撤去もこの範囲だけ消せばよい)。
CREATE TEMP TABLE _cfg AS
SELECT
  CASE :'size' WHEN 'small' THEN 100 WHEN 'medium' THEN 1000 WHEN 'large' THEN 10000 END::int  AS n_users,
  CASE :'size' WHEN 'small' THEN  10 WHEN 'medium' THEN   50 WHEN 'large' THEN   200 END::int  AS n_courses,
  CASE :'size' WHEN 'small' THEN  10 WHEN 'medium' THEN   20 WHEN 'large' THEN    50 END::int  AS chapters_per_course,
  CASE :'size' WHEN 'small' THEN  10 WHEN 'medium' THEN  100 WHEN 'large' THEN   500 END::int  AS submissions_per_user,
  CASE :'size' WHEN 'small' THEN   5 WHEN 'medium' THEN   20 WHEN 'large' THEN   100 END::int  AS notes_per_user,
  CASE :'size' WHEN 'small' THEN  30 WHEN 'medium' THEN  180 WHEN 'large' THEN   365 END::int  AS activity_days,
  1000000::bigint AS id_base;

-- 許可外の size を弾く。CASE がどれにも一致しないと NULL になり、
-- generate_series(1, NULL) が 0 行を返すため「DELETE だけが効いて何も入らない」
-- という最悪の結果になる(既存のシードを消したうえで空になる)。
-- 破壊的な DELETE が走る前にここで止める。
DO $$
BEGIN
  IF (SELECT n_users FROM _cfg) IS NULL THEN
    RAISE EXCEPTION 'size は small / medium / large のいずれかを指定してください';
  END IF;
END $$;

-- 規模の値を psql 変数へ取り込む(以降 :n_users のように埋め込んで使う)。
SELECT n_users, n_courses, chapters_per_course, submissions_per_user, notes_per_user, activity_days
  FROM _cfg \gset

\echo '=== seed-local: size =' :'size' '/ users =' :n_users '/ submissions_per_user =' :submissions_per_user

-- 依存の子から消す(FK が無くても順序は揃えておく)。
BEGIN;

DELETE FROM user_daily_activities
WHERE user_id >= 1000000;

DELETE FROM user_chapter_views
WHERE user_id >= 1000000;

DELETE FROM user_chapter_progress
WHERE user_id >= 1000000;

DELETE FROM exercise_submissions
WHERE user_id >= 1000000;

DELETE FROM notes
WHERE user_id >= 1000000;

DELETE FROM profiles
WHERE user_id >= 1000000;

DELETE FROM course_chapters
WHERE id >= 1000000;

DELETE FROM courses
WHERE id >= 1000000;

DELETE FROM master_exercises
WHERE id >= 1000000;

DELETE FROM user_oidc_identities
WHERE user_id >= 1000000;

DELETE FROM users
WHERE id >= 1000000;

-- 会社は起動時マイグレーションの seedCompanies が id=1 を入れている前提。無ければ作る。
-- 所属の正本は workspace_id なので、起動時バックフィルが会社 1 に用意したワークスペースを
-- 以降の INSERT で使う（バックフィル前に流すとワークスペース未紐付けになるため、
-- seed は必ずサーバーを 1 度起動したあとに流す）。
INSERT INTO companies (id, name)
SELECT 1, '株式会社FreStyle'
WHERE NOT EXISTS (SELECT 1 FROM companies WHERE id = 1);

-- ここで会社 1 にワークスペースが無いなら、上の INSERT で今まさに作った行か、バックフィル前の
-- DB のどちらか。そのまま進めると以降の users / courses / course_chapters が所属 NULL で入り、
-- company_id はもう無いので次回起動でも復元できない。黙って壊れた seed を作らず、ここで止める。
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM companies WHERE id = 1 AND workspace_id IS NOT NULL) THEN
    RAISE EXCEPTION '会社 1 にワークスペースが紐付いていません。サーバーを 1 度起動してバックフィルを走らせてから seed を流してください。';
  END IF;
END $$;

-- ---- users ----------------------------------------------------------------
-- 1% を company_admin にして権限分岐のあるクエリも実データで踏めるようにする。
-- ロールは role_id が正（FRESTYLE-311）。OIDC subject は user_oidc_identities に持つ。
-- password_hash は全員 'password' の bcrypt（ローカルのパスワードログイン用・FRESTYLE-311 PR2）。
INSERT INTO users (id, email, name, workspace_id, role_id, password_hash, is_active, created_at, updated_at)
SELECT
  1000000 + i,
  'seed' || i || '@example.test',
  'シード利用者' || i,
  (SELECT workspace_id FROM companies WHERE id = 1),
  (SELECT id
   FROM roles
   WHERE name = CASE WHEN i % 100 = 0 THEN 'company_admin' ELSE 'trainee' END),
  '$2a$10$Xgxiol1/CKW0E2qp4P3JOO/fZp3dcDmXxMHk76rHrOLRec8RIaqEm',
  true,
  now() - (random() * 365)::int * interval '1 day',
  now()
FROM generate_series(1, :n_users) AS i;

-- オフラインで管理画面まで触れるよう、運営管理者を 1 人入れる（admin@example.test / password）。
-- id 1000000 は連番（1000000 + i, i >= 1）と衝突しない。
INSERT INTO users (id, email, name, workspace_id, role_id, password_hash, is_active, created_at, updated_at)
VALUES (
  1000000, 'admin@example.test', 'シード運営管理者', NULL,
  (SELECT id
   FROM roles
   WHERE name = 'super_admin'),
  '$2a$10$Xgxiol1/CKW0E2qp4P3JOO/fZp3dcDmXxMHk76rHrOLRec8RIaqEm', true, now(), now()
);

-- OIDC identity（正規化後のログイン突き合わせの正）。seed の sub はダミーで、
-- Cognito ログインには使えないが、ローカルのパスワードログインが発行するトークンの sub になる。
-- 運営管理者（id 1000000）も含めて全 seed ユーザーに 1 対 1 で作る。
INSERT INTO user_oidc_identities (user_id, provider, subject, created_at, updated_at)
SELECT 1000000 + i, 'cognito', 'seed-sub-' || i, now(), now()
FROM generate_series(1, :n_users) AS i;

INSERT INTO user_oidc_identities (user_id, provider, subject, created_at, updated_at)
VALUES (1000000, 'cognito', 'seed-sub-admin', now(), now());

INSERT INTO profiles (user_id, bio, avatar_url, status_message, updated_at)
SELECT 1000000 + i, 'シード用の自己紹介文です。', '', '学習中', now()
FROM generate_series(1, :n_users) AS i;

INSERT INTO profiles (user_id, bio, avatar_url, status_message, updated_at)
VALUES (1000000, 'シード運営管理者です。', '', '運用中', now());

-- ---- courses / chapters ---------------------------------------------------
INSERT INTO courses (id, workspace_id, created_by_user_id, title, description, category, language,
                     sort_order, is_published, created_at, updated_at)
SELECT
  1000000 + c,
  (SELECT workspace_id FROM companies WHERE id = 1),
  1000000 + 1,
  'シード講座 ' || c,
  'ダミーの説明文。実教材は非公開リポが正本のためここでは扱わない。',
  (ARRAY['programming','infrastructure','database','security'])[1 + (c % 4)],
  (ARRAY['go','php','sql','bash'])[1 + (c % 4)],
  c,
  true,
  now(), now()
FROM generate_series(1, :n_courses) AS c;

INSERT INTO course_chapters (id, workspace_id, course_id, created_by_user_id, title,
                             sort_order, is_published, created_at, updated_at)
SELECT
  1000000 + ((c - 1) * :chapters_per_course + ch),
  (SELECT workspace_id FROM companies WHERE id = 1),
  1000000 + c,
  1000000 + 1,
  '第' || ch || '章',
  ch,
  true,
  now(), now()
FROM generate_series(1, :n_courses) AS c,
     generate_series(1, :chapters_per_course) AS ch;

-- ---- master_exercises -----------------------------------------------------
INSERT INTO master_exercises (id, slug, language, sort_order, category, title, description,
                              starter_code, hint_text, expected_output, mode, explanation,
                              difficulty, is_published, created_at, updated_at)
SELECT
  1000000 + e,
  'seed-exercise-' || e,
  (ARRAY['go','php','sql','bash'])[1 + (e % 4)],
  e,
  'seed',
  'シード演習 ' || e,
  'ダミーの問題文。',
  '// ここにコードを書く',
  'ヒント',
  'expected',
  'execute',
  '',
  1 + (e % 3),
  true,
  now(), now()
FROM generate_series(1, 200) AS e;

COMMIT;

-- ---- exercise_submissions -------------------------------------------------
-- 最大の行数になるテーブル。idx_submissions_user_at (user_id, submitted_at DESC) の
-- 効きを見る主対象。日時は過去 1 年に散らす。
\echo '=== exercise_submissions を投入中(最も件数が多い) ...'
BEGIN;
INSERT INTO exercise_submissions (user_id, exercise_kind, exercise_id, submitted_code,
                                  stdout, stderr, exit_code, is_correct, submitted_at)
SELECT
  1000000 + u,
  'master',
  1000000 + (1 + (s % 200)),
  'print("seed ' || s || '")',
  'seed output',
  '',
  0,
  (random() < 0.6),
  now() - (random() * 365)::int * interval '1 day' - (random() * 86400)::int * interval '1 second'
FROM generate_series(1, :n_users) AS u,
     generate_series(1, :submissions_per_user) AS s;
COMMIT;

-- ---- notes ----------------------------------------------------------------
BEGIN;
INSERT INTO notes (user_id, title, content, is_public, is_pinned, created_at, updated_at)
SELECT
  1000000 + u,
  'シードノート ' || n,
  repeat('学習メモのダミー本文。', 20),
  (random() < 0.2),
  false,
  now() - (random() * 365)::int * interval '1 day',
  now()
FROM generate_series(1, :n_users) AS u,
     generate_series(1, :notes_per_user) AS n;
COMMIT;

-- ---- 学習の進捗 / 閲覧 / 日次集計 -------------------------------------------
-- ダッシュボードが読む 3 テーブル。各利用者が最初の 3 講座を進めている想定にする。
\echo '=== 進捗 / 閲覧 / 日次集計を投入中 ...'
BEGIN;
INSERT INTO user_chapter_progress (user_id, chapter_id, course_id, completed_at, created_at)
SELECT
  1000000 + u,
  1000000 + ((c - 1) * :chapters_per_course + ch),
  1000000 + c,
  now() - (random() * 180)::int * interval '1 day',
  now()
FROM generate_series(1, :n_users) AS u,
     generate_series(1, LEAST(3, :n_courses)) AS c,
     generate_series(1, :chapters_per_course) AS ch
WHERE random() < 0.5
ON CONFLICT DO NOTHING;

INSERT INTO user_chapter_views (user_id, chapter_id, course_id, first_viewed_at, last_viewed_at, view_count)
SELECT
  1000000 + u,
  1000000 + ((c - 1) * :chapters_per_course + ch),
  1000000 + c,
  v.first_viewed_at,
  -- 初回閲覧から now() までの間で最終閲覧を決める。両者を独立に生成すると
  -- last < first の行ができ、閲覧期間を扱う集計が壊れる。
  v.first_viewed_at
    + (random() * extract(epoch FROM now() - v.first_viewed_at))::int * interval '1 second',
  1 + (random() * 9)::int
FROM generate_series(1, :n_users) AS u,
     generate_series(1, LEAST(3, :n_courses)) AS c,
     generate_series(1, :chapters_per_course) AS ch,
     LATERAL (SELECT now() - (random() * 180)::int * interval '1 day' AS first_viewed_at) AS v
WHERE random() < 0.7
ON CONFLICT DO NOTHING;

INSERT INTO user_daily_activities (user_id, activity_date, exercise_count, correct_count,
                                   chapter_count, note_count)
SELECT
  1000000 + u,
  (now() - d * interval '1 day')::date,
  e.exercise_count,
  -- 正答数は提出数を超えさせない。独立に生成すると正答率が 100% を超える行ができ、
  -- ダッシュボードの集計が破綻する。
  (random() * e.exercise_count)::int,
  (random() * 3)::int,
  (random() * 2)::int
FROM generate_series(1, :n_users) AS u,
     generate_series(0, :activity_days - 1) AS d,
     LATERAL (SELECT (random() * 10)::int AS exercise_count) AS e
WHERE random() < 0.4
ON CONFLICT DO NOTHING;
COMMIT;

-- ---- 統計の更新 ------------------------------------------------------------
-- ANALYZE を忘れるとプランナが古い統計で判断し、実行計画の比較が無意味になる。
\echo '=== ANALYZE 実行中 ...'
ANALYZE users, profiles, courses, course_chapters, master_exercises,
        exercise_submissions, notes, user_chapter_progress, user_chapter_views,
        user_daily_activities;

-- 規模の受け渡しに使った一時テーブルは、この後の集計に混ざらないよう捨てる。
DROP TABLE _cfg;

-- ---- 結果 ------------------------------------------------------------------
\echo ''
\echo '=== 投入結果 ==='
SELECT relname AS table_name,
       to_char(n_live_tup, 'FM999,999,999') AS rows,
       pg_size_pretty(pg_total_relation_size(relid)) AS total_size
  FROM pg_stat_user_tables
 WHERE n_live_tup > 0
 ORDER BY n_live_tup DESC;
