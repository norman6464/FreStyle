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
-- courses / course_chapters は全廃済みのテーブルなので n_courses / chapters_per_course は
-- 持たない(かつては courses / course_chapters / user_chapter_progress / user_chapter_views の
-- 規模指定に使っていたが、いずれも現行 schema.hcl に存在しない)。
CREATE TEMP TABLE _cfg AS
SELECT
  CASE :'size' WHEN 'small' THEN 100 WHEN 'medium' THEN 1000 WHEN 'large' THEN 10000 END::int  AS n_users,
  CASE :'size' WHEN 'small' THEN  10 WHEN 'medium' THEN  100 WHEN 'large' THEN   500 END::int  AS submissions_per_user,
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
SELECT n_users, submissions_per_user, activity_days
  FROM _cfg \gset

\echo '=== seed-local: size =' :'size' '/ users =' :n_users '/ submissions_per_user =' :submissions_per_user

-- 依存の子から消す(FK が無くても順序は揃えておく)。
BEGIN;

DELETE FROM user_daily_activities
WHERE user_id >= 1000000;

DELETE FROM exercise_submissions
WHERE user_id >= 1000000;

DELETE FROM profiles
WHERE user_id >= 1000000;

DELETE FROM master_exercises
WHERE id >= 1000000;

DELETE FROM user_oidc_identities
WHERE user_id >= 1000000;

DELETE FROM users
WHERE id >= 1000000;

-- ---- users ----------------------------------------------------------------
-- companies / roles テーブルは会社→ワークスペース移行のレガシー橋渡し撤去(#2413)で
-- 全廃済みなので、もう company_id 経由で workspace_id を引けないし role_id 列自体も無い。
-- 運営管理者かどうかは DB の列ではなく、ログイン時に発行される OIDC トークンの
-- groups クレーム(OIDC_ROLES_CLAIM。config.go の AdminRoleClaim/AdminRole)に
-- "admin" が入っているかで決まるため、ここでは判定材料を持たせようがない
-- (かつ持たせる必要も無い)。workspace_id は所属先が無いので NULL のままにする
-- (users.workspace_id は nullable)。
-- password_hash も同じ理由で投入しない: パスワード検証はもう DB 側ではなく
-- Dex(docker/idp/config.yaml の staticPasswords)側が持つ。列自体は残っているので
-- INSERT の列リストから外し、NULL のまま(既定値は無い)にしている。
INSERT INTO users (id, email, name, workspace_id, is_active, created_at, updated_at)
SELECT
  1000000 + i,
  'seed' || i || '@example.test',
  'シード利用者' || i,
  NULL,
  true,
  now() - (random() * 365)::int * interval '1 day',
  now()
FROM generate_series(1, :n_users) AS i;

-- オフラインで管理画面まで触れるよう、運営管理者を 1 人入れる
-- (admin@example.test / password。Dex 側の docker/idp/config.yaml staticPasswords で認証する)。
-- id 1000000 は連番（1000000 + i, i >= 1）と衝突しない。
INSERT INTO users (id, email, name, workspace_id, is_active, created_at, updated_at)
VALUES (
  1000000, 'admin@example.test', 'シード運営管理者', NULL, true, now(), now()
);

-- OIDC identity（正規化後のログイン突き合わせの正）。
--
-- bulk の seed1..N@example.test には Dex 側に対応する staticPasswords が無く、実際には
-- 誰もログインしない(exercise_submissions 等のダミーデータ量産のためだけに存在する)ので、
-- subject はダミー文字列のままでよい。provider は "cognito" 固定
-- (domain.OidcProviderCognito。歴史的な名残りで実際の発行者を指す値ではないが、
-- FindByCognitoSub 等がこの文字列で照合するため、発行者を Dex に変えても値は変えない)。
INSERT INTO user_oidc_identities (user_id, provider, subject, created_at, updated_at)
SELECT 1000000 + i, 'cognito', 'seed-sub-' || i, now(), now()
FROM generate_series(1, :n_users) AS i;

-- 運営管理者(id 1000000)は実際に Dex でログインするため、Dex が本当に発行する sub と
-- 一致させる必要がある。Dex の password connector が返す sub は、
-- userID(docker/idp/config.yaml の staticPasswords[].userID)とコネクタ名の固定値
-- "local" を internal.IDTokenSubject 相当の protobuf メッセージ
-- (field 1 = userID, field 2 = connector id)に詰めて base64url(パディング無し)した値
-- になる(公式ドキュメントには明記が無いため、実際に Dex が発行した id_token をデコードして
-- 実測・確認済み)。以下は生の protobuf バイト列を手で組み立てて同じ値を再現している:
--   \x0a <userID の長さ(1 byte)> <userID>  -- タグ 0x0a = field 1, wiretype 2(length-delimited)
--   \x12 <"local" の長さ(1 byte)> local    -- タグ 0x12 = field 2, wiretype 2
-- 長さを 1 byte(set_byte)で埋めているため、userID / "local" が 128 byte 未満
-- (protobuf のごく短い varint 長が 1 byte に収まる範囲)であることが前提
-- (このユースケースでは常に成立する)。
-- 標準 base64 の '+' '/' を '-' '_' に translate し、'=' パディングを rtrim で落として
-- base64url 化している。
--
-- ここで使う userID('seed-sub-admin')は docker/idp/config.yaml の
-- staticPasswords[].userID と完全に一致させること。ずれると、Dex が発行する sub が
-- ここに登録した subject と一致せず、ログイン時に「未知の sub」として
-- 新規サインアップが走り、既に使われている admin@example.test で 409 email_taken になる。
INSERT INTO user_oidc_identities (user_id, provider, subject, created_at, updated_at)
SELECT
  1000000,
  'cognito',
  rtrim(translate(encode(
    '\x0a'::bytea || set_byte('\x00'::bytea, 0, octet_length(u)) || convert_to(u, 'UTF8') ||
    '\x12'::bytea || set_byte('\x00'::bytea, 0, octet_length('local')) || convert_to('local', 'UTF8'),
    'base64'
  ), E'+/\n', '-_'), '='),
  now(),
  now()
FROM (SELECT 'seed-sub-admin'::text AS u) AS dex_local_subject;

INSERT INTO profiles (user_id, bio, avatar_url, status_message, updated_at)
SELECT 1000000 + i, 'シード用の自己紹介文です。', '', '学習中', now()
FROM generate_series(1, :n_users) AS i;

INSERT INTO profiles (user_id, bio, avatar_url, status_message, updated_at)
VALUES (1000000, 'シード運営管理者です。', '', '運用中', now());

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

-- ---- 学習の日次集計 ---------------------------------------------------------
-- ダッシュボードが読むテーブル。course_chapters が全廃済みのため、かつてここにあった
-- user_chapter_progress / user_chapter_views(講座の章単位の進捗・閲覧)への投入は
-- 対象テーブルごと無くなっている。
\echo '=== 日次集計を投入中 ...'
BEGIN;
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
ANALYZE users, profiles, master_exercises, exercise_submissions, user_daily_activities;

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
