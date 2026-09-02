-- 教材（コース / 章）の権限。
--
-- ノートと同じ「事実だけを集めて、規則は domain が持つ」分担にする。ここが返すのは
-- 所属・公開状態・ワークスペースの admin か・届いている役割の強さの 4 つで、
-- そこから何ができるかは domain.ResolveMaterialPermission だけが決める。
--
-- 役割は強さ（整数）で返す。役割そのもの（text）を返すと「付与が 1 つも無い」が NULL に
-- なり、生成コードの型付けが崩れる（ノート側と同じ理由）。

-- name: ResolveChapterPermissionFacts :one
-- 章 1 つの実効権限を決める事実を 1 回のクエリで集める。
--
-- **ワークスペースの grant は admin だけを見る。** editor / commenter / viewer は
-- ノートの木に対する既定で、教材には届かせない（届かせると、ノートの editor である人が
-- 教材の編集権まで一度に得てしまう）。この 1 行がその境目そのもの。
--
-- 「スペースの全員」（kind='space_all'）も混ぜない。あれはノートのスペースに紐づく主体で、
-- 教材はスペースに属さない。
WITH me AS (
    SELECT pr.id
    FROM principals pr
    WHERE pr.workspace_id = sqlc.arg(workspace_id)
      AND pr.kind = 'user' AND pr.user_id = sqlc.arg(user_id)
),
mine AS (
    SELECT id FROM me
    UNION
    SELECT pm.group_principal_id
    FROM principal_members pm
    JOIN me ON me.id = pm.member_principal_id
    WHERE pm.workspace_id = sqlc.arg(workspace_id)
),
target AS (
    SELECT ch.id, ch.course_id, ch.is_published
    FROM course_chapters ch
    WHERE ch.workspace_id = sqlc.arg(workspace_id) AND ch.id = sqlc.arg(chapter_id)
)
SELECT
    EXISTS (SELECT 1 FROM target) AS target_exists,
    EXISTS (SELECT 1 FROM me) AS is_member,
    -- COALESCE(bool 式, false) だと sqlc の型推論が interface{} に落ちるので EXISTS で書く
    -- （driver が NULL を bool へ Scan できずに落ちることを別の口で確認済み）。
    -- 対象が無ければ false になるが、その場合は呼び出し側が target_exists で先に断る。
    EXISTS (SELECT 1 FROM target WHERE is_published) AS is_published,
    EXISTS (
        SELECT 1 FROM workspace_grants wg
         WHERE wg.workspace_id = sqlc.arg(workspace_id)
           AND wg.principal_id IN (SELECT id FROM mine)
           AND wg."role" = 'admin'
    ) AS is_workspace_admin,
    GREATEST(
        COALESCE((
            SELECT max(CASE cg."role"
                         WHEN 'admin' THEN 4 WHEN 'editor' THEN 3
                         WHEN 'commenter' THEN 2 WHEN 'viewer' THEN 1 ELSE 0 END)
            FROM course_grants cg
            JOIN target t ON t.course_id = cg.course_id
            WHERE cg.workspace_id = sqlc.arg(workspace_id)
              AND cg.principal_id IN (SELECT id FROM mine)
        ), 0),
        COALESCE((
            SELECT max(CASE chg."role"
                         WHEN 'admin' THEN 4 WHEN 'editor' THEN 3
                         WHEN 'commenter' THEN 2 WHEN 'viewer' THEN 1 ELSE 0 END)
            FROM chapter_grants chg
            JOIN target t ON t.id = chg.chapter_id
            WHERE chg.workspace_id = sqlc.arg(workspace_id)
              AND chg.principal_id IN (SELECT id FROM mine)
        ), 0)
    )::integer AS grant_rank;

-- name: ResolveCoursePermissionFacts :one
-- コース 1 つの実効権限を決める事実を集める。見方は章と同じで、章の付与は見ない
-- （章に張った権限はその章にしか効かず、コースそのものへは上がらない）。
WITH me AS (
    SELECT pr.id
    FROM principals pr
    WHERE pr.workspace_id = sqlc.arg(workspace_id)
      AND pr.kind = 'user' AND pr.user_id = sqlc.arg(user_id)
),
mine AS (
    SELECT id FROM me
    UNION
    SELECT pm.group_principal_id
    FROM principal_members pm
    JOIN me ON me.id = pm.member_principal_id
    WHERE pm.workspace_id = sqlc.arg(workspace_id)
),
target AS (
    SELECT c.id, c.is_published
    FROM courses c
    WHERE c.workspace_id = sqlc.arg(workspace_id) AND c.id = sqlc.arg(course_id)
)
SELECT
    EXISTS (SELECT 1 FROM target) AS target_exists,
    EXISTS (SELECT 1 FROM me) AS is_member,
    -- COALESCE(bool 式, false) だと sqlc の型推論が interface{} に落ちるので EXISTS で書く
    -- （driver が NULL を bool へ Scan できずに落ちることを別の口で確認済み）。
    -- 対象が無ければ false になるが、その場合は呼び出し側が target_exists で先に断る。
    EXISTS (SELECT 1 FROM target WHERE is_published) AS is_published,
    EXISTS (
        SELECT 1 FROM workspace_grants wg
         WHERE wg.workspace_id = sqlc.arg(workspace_id)
           AND wg.principal_id IN (SELECT id FROM mine)
           AND wg."role" = 'admin'
    ) AS is_workspace_admin,
    COALESCE((
        SELECT max(CASE cg."role"
                     WHEN 'admin' THEN 4 WHEN 'editor' THEN 3
                     WHEN 'commenter' THEN 2 WHEN 'viewer' THEN 1 ELSE 0 END)
        FROM course_grants cg
        JOIN target t ON t.id = cg.course_id
        WHERE cg.workspace_id = sqlc.arg(workspace_id)
          AND cg.principal_id IN (SELECT id FROM mine)
    ), 0)::integer AS grant_rank;

-- name: UpsertCourseGrant :one
-- コースでの既定の役割の付与（同じ主体には 1 行だけ）。
INSERT INTO course_grants (workspace_id, course_id, principal_id, "role")
VALUES ($1, $2, $3, $4)
ON CONFLICT (workspace_id, course_id, principal_id)
DO UPDATE SET "role" = EXCLUDED."role", updated_at = now()
RETURNING *;

-- name: DeleteCourseGrant :execrows
DELETE FROM course_grants
WHERE workspace_id = $1 AND course_id = $2 AND principal_id = $3;

-- name: ListCourseGrants :many
-- そのコース自身に張られた付与の一覧（章に張った分は含まない）。
SELECT * FROM course_grants
WHERE workspace_id = $1 AND course_id = $2
ORDER BY principal_id;

-- name: UpsertChapterGrant :one
-- 章 1 つでの既定の役割の付与。コースの付与より弱い役割をここに張っても下がらない
-- （合成は最も強いものを採る）。
INSERT INTO chapter_grants (workspace_id, chapter_id, principal_id, "role")
VALUES ($1, $2, $3, $4)
ON CONFLICT (workspace_id, chapter_id, principal_id)
DO UPDATE SET "role" = EXCLUDED."role", updated_at = now()
RETURNING *;

-- name: DeleteChapterGrant :execrows
DELETE FROM chapter_grants
WHERE workspace_id = $1 AND chapter_id = $2 AND principal_id = $3;

-- name: ListChapterGrants :many
-- その章自身に張られた付与の一覧（コースから降りてくる分は含まない）。
SELECT * FROM chapter_grants
WHERE workspace_id = $1 AND chapter_id = $2
ORDER BY principal_id;

-- name: ListCourseFactsForUser :many
-- ワークスペース内のコース全件と、その実効権限を決める事実を 1 回のクエリで返す。
--
-- **返り値はまだ「見せてよいコース」ではない。** ふるい落とすのは呼び出し側で、
-- 判定は domain.ResolveMaterialPermission が行う（ここで絞ると規則が 2 箇所に分かれる）。
--
-- コースごとに ResolveCoursePermissionFacts を呼ぶ（N+1）ことはしない。一覧は画面を
-- 開くたびに引くので、コース数だけ往復するとそのまま待ち時間になる。
WITH me AS (
    SELECT pr.id
    FROM principals pr
    WHERE pr.workspace_id = sqlc.arg(workspace_id)
      AND pr.kind = 'user' AND pr.user_id = sqlc.arg(user_id)
),
mine AS (
    SELECT id FROM me
    UNION
    SELECT pm.group_principal_id
    FROM principal_members pm
    JOIN me ON me.id = pm.member_principal_id
    WHERE pm.workspace_id = sqlc.arg(workspace_id)
),
ranks AS (
    SELECT cg.course_id,
           max(CASE cg."role"
                 WHEN 'admin' THEN 4 WHEN 'editor' THEN 3
                 WHEN 'commenter' THEN 2 WHEN 'viewer' THEN 1 ELSE 0 END) AS rank
    FROM course_grants cg
    WHERE cg.workspace_id = sqlc.arg(workspace_id)
      AND cg.principal_id IN (SELECT id FROM mine)
    GROUP BY cg.course_id
)
SELECT c.*,
    EXISTS (SELECT 1 FROM me) AS is_member,
    EXISTS (
        SELECT 1 FROM workspace_grants wg
         WHERE wg.workspace_id = sqlc.arg(workspace_id)
           AND wg.principal_id IN (SELECT id FROM mine)
           AND wg."role" = 'admin'
    ) AS is_workspace_admin,
    COALESCE(r.rank, 0)::integer AS grant_rank
FROM courses c
LEFT JOIN ranks r ON r.course_id = c.id
WHERE c.workspace_id = sqlc.arg(workspace_id)
ORDER BY c.sort_order, c.id;
