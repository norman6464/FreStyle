-- name: ListMasterExercisesByLanguage :many
-- 公開済みの問題を返す。language が空文字なら全言語。
-- sort_order は一意でない（既定値 0）。同値行の相対順序は SQL 上未定義なので id をタイブレークに置く。
SELECT id, slug, language, sort_order, category, title, description, starter_code,
       hint_text, expected_output, mode, explanation, difficulty, is_published,
       chapter_id, created_at, updated_at
FROM master_exercises
WHERE is_published = TRUE
  AND (sqlc.arg(language)::text = '' OR language = sqlc.arg(language)::text)
ORDER BY sort_order ASC, id ASC;

-- name: GetMasterExerciseByID :one
-- 内部 ID で 1 件取得（存在しなければ sql.ErrNoRows）。非公開でも返す。
SELECT id, slug, language, sort_order, category, title, description, starter_code,
       hint_text, expected_output, mode, explanation, difficulty, is_published,
       chapter_id, created_at, updated_at
FROM master_exercises
WHERE id = sqlc.arg(id);

-- name: GetMasterExerciseBySlug :one
-- slug で 1 件取得（存在しなければ sql.ErrNoRows）。slug は一意（uq）なので最大 1 行。
SELECT id, slug, language, sort_order, category, title, description, starter_code,
       hint_text, expected_output, mode, explanation, difficulty, is_published,
       chapter_id, created_at, updated_at
FROM master_exercises
WHERE slug = sqlc.arg(slug);

-- name: SummarizeMasterExercisesByLanguage :many
-- 公開済み問題を言語ごとに集計し、問題数と current user の正解済み件数を返す。
-- 言語選択カード用で、問題本文を返さないので一覧より軽い。
-- viewer_id=0（未ログイン）は usr サブクエリが空になり solved は全て 0。
-- 正解は「問題ごとに 1 回でも正解したか」なので BOOL_OR で問題単位に畳んでから数える。
SELECT e.language                                     AS language,
       COUNT(*)                                       AS total,
       COUNT(*) FILTER (WHERE usr.any_solved IS TRUE) AS solved
FROM master_exercises e
LEFT JOIN (
    SELECT su.exercise_id, BOOL_OR(su.is_correct) AS any_solved
    FROM exercise_submissions su
    WHERE su.exercise_kind = sqlc.arg(exercise_kind)::text
      AND su.user_id = sqlc.arg(viewer_id)
    GROUP BY su.exercise_id
) usr ON usr.exercise_id = e.id
WHERE e.is_published = TRUE
GROUP BY e.language
ORDER BY e.language ASC;

-- name: ListMasterExercisesWithStatusByLanguage :many
-- 公開済み問題を「current user の提出状態 + 全体集計」付きで 1 クエリで返す。
-- viewer_id=0（未ログイン）は usr サブクエリが空になり status は全て ''。
--
-- sort_order は一意でないため id をタイブレークに置く。これが無いと同値行の順序がページ間で
-- 揺れ、OFFSET ページングで同じ行が重複したり抜け落ちたりする。
--
-- ページングは row_limit=0 を「全件」として扱う。PostgreSQL の LIMIT NULL は無制限なので
-- NULLIF で 0 を NULL に倒す（クエリを 2 本に分けずに済む）。row_limit / row_offset は
-- bigint で受ける（LIMIT / OFFSET の型に合わせ、Go 側で int32 へ縮めないため）。
SELECT e.id, e.slug, e.language, e.sort_order, e.category, e.title, e.description,
       e.starter_code, e.hint_text, e.expected_output, e.mode, e.explanation,
       e.difficulty, e.is_published, e.chapter_id, e.created_at, e.updated_at,
       COALESCE(agg.total_submissions, 0)::bigint AS total_submissions,
       COALESCE(agg.solved_users, 0)::bigint      AS solved_users,
       (CASE
           WHEN usr.any_solved IS TRUE  THEN 'solved'
           WHEN usr.any_solved IS FALSE THEN 'in_progress'
           ELSE ''
       END)::text AS user_status
FROM master_exercises e
LEFT JOIN (
    SELECT sa.exercise_id,
           COUNT(*)                                             AS total_submissions,
           COUNT(DISTINCT sa.user_id) FILTER (WHERE sa.is_correct) AS solved_users
    FROM exercise_submissions sa
    WHERE sa.exercise_kind = sqlc.arg(exercise_kind)::text
    GROUP BY sa.exercise_id
) agg ON agg.exercise_id = e.id
LEFT JOIN (
    SELECT su.exercise_id, BOOL_OR(su.is_correct) AS any_solved
    FROM exercise_submissions su
    WHERE su.exercise_kind = sqlc.arg(exercise_kind)::text
      AND su.user_id = sqlc.arg(viewer_id)
    GROUP BY su.exercise_id
) usr ON usr.exercise_id = e.id
WHERE e.is_published = TRUE
  AND (sqlc.arg(language)::text = '' OR e.language = sqlc.arg(language)::text)
ORDER BY e.sort_order ASC, e.id ASC
LIMIT NULLIF(sqlc.arg(row_limit)::bigint, 0)
OFFSET sqlc.arg(row_offset)::bigint;
