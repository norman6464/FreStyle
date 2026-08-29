-- 0024 の切り戻し — 統合した中身を、元の個人ワークスペースへ戻す。
--
-- 【前提】
--   0024_merge_personal_workspaces_into_company.sql が作った kb_merge_backup スキーマの
--   控えを使う。控えを消したあと（DROP SCHEMA kb_merge_backup CASCADE）は戻せない。
--
-- 【何をどこまで戻すか】
--   - 移したスペースは、控えに記録した元のワークスペースへ戻す。
--   - **統合後に増えた行も、入れ物ごと一緒に戻す**（移したスペースの中に作られたページ、
--     そのページのブロック・制限・公開リンクなど）。ページはスペースに属するものなので、
--     スペースが戻るなら中身も戻るのが筋。ここで「控えにある行だけ」を戻すと、
--     残された子行が消えた親を指すことになり、FK でどのみち通らない。
--   - 統合先で新しく作られたスペース・その配下は動かさない。
--   - 統合で作った統合先の主体（principals）と既定の役割は、他から参照されていなければ消す。
--     参照が残っているものは残して NOTICE で報告する（会社ワークスペースのメンバーが
--     1 人増えたままになるだけで、実害は無い）。
--
-- 【戻せないとき】
--   統合後に「その個人ワークスペースには居なかった人」へスペース／ページの権限を与えていると、
--   戻し先に対応する主体が無い。黙って消すと権限設定が失われるので、**先に止めて報告する**。
--   報告された行を消すか、その人を残す判断をしてから、もう一度この SQL を流す。
--
-- 【FK の都合】
--   移行本体と同じ理由（複合 FK・DEFERRABLE でない）で、テナントを戻す UPDATE は 1 文にまとめる。

BEGIN;

SET LOCAL lock_timeout = '30s';
SET LOCAL statement_timeout = '10min';

-- ============================================================================
-- 0. 戻す対象の集合を先に固める（この後の書き換えに引きずられないよう実体化する）
-- ============================================================================

-- 移したスペース → 戻し先のワークスペース
CREATE TEMP TABLE _rb_space ON COMMIT DROP AS
SELECT b.id AS space_id, b.workspace_id AS src_ws
  FROM kb_merge_backup.spaces b
  JOIN kb_merge_backup.plan p ON p.source_workspace_id = b.workspace_id AND p.merged_at IS NOT NULL;

-- 移したスペースの下に今あるページ（統合後に作られたページも含む）
CREATE TEMP TABLE _rb_page ON COMMIT DROP AS
SELECT x.id AS page_id, s.src_ws
  FROM pages x JOIN _rb_space s ON s.space_id = x.space_id;

-- 戻したあと、その主体が戻し先のワークスペースに実在するか（権限行の検査に使う）
CREATE TEMP TABLE _rb_principal_in_source ON COMMIT DROP AS
          SELECT m.target_principal_id AS principal_id, m.source_workspace_id AS src_ws
            FROM kb_merge_backup.principal_map m
UNION ALL SELECT b.id, b.workspace_id
            FROM kb_merge_backup.principals b WHERE b.kind = 'group'
UNION ALL SELECT p.id, s.src_ws
            FROM principals p JOIN _rb_space s ON s.space_id = p.space_id WHERE p.kind = 'space_all'
UNION ALL SELECT p.id, g.src_ws
            FROM principals p JOIN _rb_page g ON g.page_id = p.page_id WHERE p.kind = 'share_link';

CREATE INDEX ON _rb_space (space_id);
CREATE INDEX ON _rb_page (page_id);
CREATE INDEX ON _rb_principal_in_source (principal_id, src_ws);

DO $target$
DECLARE
    n bigint;
BEGIN
    SELECT count(*) INTO n FROM kb_merge_backup.plan WHERE merged_at IS NOT NULL;
    IF n = 0 THEN
        RAISE EXCEPTION '戻す対象がない（kb_merge_backup.plan に適用済みの行が 1 件も無い）';
    END IF;
    RAISE NOTICE '切り戻し対象: ワークスペース % 個 / スペース % 個 / ページ % 枚',
        n, (SELECT count(*) FROM _rb_space), (SELECT count(*) FROM _rb_page);
END
$target$;

-- ============================================================================
-- 1. 戻せない状態でないかを先に確かめる（統合後に足された権限で、戻し先に主体が無いもの）
-- ============================================================================
DO $blockers$
DECLARE
    detail text := '';
    part   text;
BEGIN
    SELECT string_agg(format(E'\n  - space_grants: space=%s principal=%s role=%s', x.space_id, x.principal_id, x."role"), '')
      INTO part
      FROM space_grants x
      JOIN _rb_space s ON s.space_id = x.space_id
     WHERE NOT EXISTS (SELECT 1 FROM kb_merge_backup.created_space_grant c
                        WHERE c.workspace_id = x.workspace_id AND c.space_id = x.space_id
                          AND c.principal_id = x.principal_id)
       AND NOT EXISTS (SELECT 1 FROM _rb_principal_in_source r
                        WHERE r.principal_id = x.principal_id AND r.src_ws = s.src_ws);
    detail := detail || COALESCE(part, '');

    SELECT string_agg(format(E'\n  - page_restrictions: page=%s principal=%s %s/%s', x.page_id, x.principal_id, x.capability, x.mode), '')
      INTO part
      FROM page_restrictions x
      JOIN _rb_page g ON g.page_id = x.page_id
     WHERE NOT EXISTS (SELECT 1 FROM _rb_principal_in_source r
                        WHERE r.principal_id = x.principal_id AND r.src_ws = g.src_ws);
    detail := detail || COALESCE(part, '');

    SELECT string_agg(format(E'\n  - principal_members: group=%s member=%s', x.group_principal_id, x.member_principal_id), '')
      INTO part
      FROM principal_members x
      JOIN kb_merge_backup.principals b ON b.id = x.group_principal_id AND b.kind = 'group'
     WHERE NOT EXISTS (SELECT 1 FROM _rb_principal_in_source r
                        WHERE r.principal_id = x.member_principal_id AND r.src_ws = b.workspace_id);
    detail := detail || COALESCE(part, '');

    IF detail <> '' THEN
        RAISE EXCEPTION
            E'統合後に足された権限のうち、戻し先のワークスペースに主体が無いものがある。消すか残すかを決めてから流し直すこと:%',
            detail;
    END IF;
END
$blockers$;

-- ============================================================================
-- 2. 移行が作った space_grants を消す（戻したスペースに、無かった役割を残さない）
-- ============================================================================
DELETE FROM space_grants x
 USING kb_merge_backup.created_space_grant c
 WHERE x.workspace_id = c.workspace_id AND x.space_id = c.space_id AND x.principal_id = c.principal_id;

-- ============================================================================
-- 3. 捨てた source の user 主体を戻す（4 の 1 文が参照を付け替える先になる）
--    ここは単独で成立する（ワークスペースは消していない・users も居る）。
-- ============================================================================
INSERT INTO principals (id, workspace_id, kind, user_id, space_id, page_id, name, created_at, updated_at)
SELECT b.id, b.workspace_id, b.kind, b.user_id, b.space_id, b.page_id, b.name, b.created_at, b.updated_at
  FROM kb_merge_backup.principals b
 WHERE b.kind = 'user'
   AND NOT EXISTS (SELECT 1 FROM principals p WHERE p.id = b.id);

-- ============================================================================
-- 4. 本体（逆向き）— テナントを戻す。移行本体と同じ理由で **必ず 1 文**。
--    principal_id は「統合先の主体 → その人の source 側の主体」へ写像し直す。
-- ============================================================================
WITH
mv_spaces AS (
    UPDATE spaces x SET workspace_id = s.src_ws
      FROM _rb_space s WHERE x.id = s.space_id RETURNING 1
),
mv_pages AS (
    UPDATE pages x SET workspace_id = g.src_ws
      FROM _rb_page g WHERE x.id = g.page_id RETURNING 1
),
mv_blocks AS (
    UPDATE blocks x SET workspace_id = g.src_ws
      FROM _rb_page g WHERE x.page_id = g.page_id RETURNING 1
),
mv_page_paths AS (
    UPDATE page_paths x SET workspace_id = g.src_ws
      FROM _rb_page g WHERE x.page_id = g.page_id RETURNING 1
),
mv_page_allow_lists AS (
    UPDATE page_allow_lists x SET workspace_id = g.src_ws
      FROM _rb_page g WHERE x.page_id = g.page_id RETURNING 1
),
mv_page_restrictions AS (
    UPDATE page_restrictions x
       SET workspace_id = g.src_ws,
           principal_id = COALESCE((SELECT m.source_principal_id FROM kb_merge_backup.principal_map m
                                     WHERE m.target_principal_id = x.principal_id
                                       AND m.source_workspace_id = g.src_ws), x.principal_id)
      FROM _rb_page g WHERE x.page_id = g.page_id RETURNING 1
),
mv_space_grants AS (
    UPDATE space_grants x
       SET workspace_id = s.src_ws,
           principal_id = COALESCE((SELECT m.source_principal_id FROM kb_merge_backup.principal_map m
                                     WHERE m.target_principal_id = x.principal_id
                                       AND m.source_workspace_id = s.src_ws), x.principal_id)
      FROM _rb_space s WHERE x.space_id = s.space_id RETURNING 1
),
mv_share_links AS (
    UPDATE share_links x SET workspace_id = g.src_ws
      FROM _rb_page g WHERE x.page_id = g.page_id RETURNING 1
),
-- 主体は種類ごとに戻し先の手がかりが違う: group は控え、space_all はスペース、share_link はページ。
mv_principals_group AS (
    UPDATE principals x SET workspace_id = b.workspace_id
      FROM kb_merge_backup.principals b
     WHERE b.kind = 'group' AND x.id = b.id RETURNING 1
),
mv_principals_space_all AS (
    UPDATE principals x SET workspace_id = s.src_ws
      FROM _rb_space s WHERE x.kind = 'space_all' AND x.space_id = s.space_id RETURNING 1
),
mv_principals_share_link AS (
    UPDATE principals x SET workspace_id = g.src_ws
      FROM _rb_page g WHERE x.kind = 'share_link' AND x.page_id = g.page_id RETURNING 1
),
mv_principal_members AS (
    UPDATE principal_members x
       SET workspace_id        = b.workspace_id,
           member_principal_id = COALESCE((SELECT m.source_principal_id FROM kb_merge_backup.principal_map m
                                            WHERE m.target_principal_id = x.member_principal_id
                                              AND m.source_workspace_id = b.workspace_id),
                                          x.member_principal_id)
      FROM kb_merge_backup.principals b
     WHERE b.kind = 'group' AND x.group_principal_id = b.id RETURNING 1
)
SELECT (SELECT count(*) FROM mv_spaces)             AS spaces,
       (SELECT count(*) FROM mv_pages)              AS pages,
       (SELECT count(*) FROM mv_blocks)             AS blocks,
       (SELECT count(*) FROM mv_page_paths)         AS page_paths,
       (SELECT count(*) FROM mv_page_allow_lists)   AS page_allow_lists,
       (SELECT count(*) FROM mv_page_restrictions)  AS page_restrictions,
       (SELECT count(*) FROM mv_space_grants)       AS space_grants,
       (SELECT count(*) FROM mv_share_links)        AS share_links,
       (SELECT count(*) FROM mv_principals_group)
         + (SELECT count(*) FROM mv_principals_space_all)
         + (SELECT count(*) FROM mv_principals_share_link) AS principals,
       (SELECT count(*) FROM mv_principal_members)  AS principal_members;

-- ============================================================================
-- 5. 消した source の workspace_grants を戻す（主体が source 側に揃ってから）
-- ============================================================================
INSERT INTO workspace_grants (workspace_id, principal_id, "role", created_at, updated_at)
SELECT b.workspace_id, b.principal_id, b."role", b.created_at, b.updated_at
  FROM kb_merge_backup.workspace_grants b
 WHERE NOT EXISTS (SELECT 1 FROM workspace_grants g
                    WHERE g.workspace_id = b.workspace_id AND g.principal_id = b.principal_id);

-- ============================================================================
-- 6. 統合先に作った主体と役割を片付ける（他から参照されていないものだけ）
-- ============================================================================
DELETE FROM workspace_grants x
 USING kb_merge_backup.created_workspace_grant c
 WHERE x.workspace_id = c.workspace_id AND x.principal_id = c.principal_id AND x."role" = c."role";

DELETE FROM principals p
 USING kb_merge_backup.created_principal c
 WHERE p.id = c.id
   AND NOT EXISTS (SELECT 1 FROM workspace_grants  g  WHERE g.principal_id  = p.id)
   AND NOT EXISTS (SELECT 1 FROM space_grants      sg WHERE sg.principal_id = p.id)
   AND NOT EXISTS (SELECT 1 FROM page_restrictions r  WHERE r.principal_id  = p.id)
   AND NOT EXISTS (SELECT 1 FROM principal_members m  WHERE m.member_principal_id = p.id);

DO $left$
DECLARE
    row_text text;
BEGIN
    SELECT string_agg(format(E'\n  - principal %s（user %s / workspace %s）', c.id, c.user_id, c.workspace_id), '')
      INTO row_text
      FROM kb_merge_backup.created_principal c
      JOIN principals p ON p.id = c.id;
    IF row_text IS NOT NULL THEN
        RAISE NOTICE
            E'統合で作った主体のうち、統合後に別の権限が張られたものは残した（会社ワークスペースのメンバーとして残るだけ）:%',
            row_text;
    END IF;
END
$left$;

-- ============================================================================
-- 7. 衝突回避で改名した値を戻す
-- ============================================================================
-- updated_at も控えの値へ戻す（now() を入れると「移行前とまったく同じ」でなくなる）
UPDATE spaces s SET "key" = r.old_value, updated_at = b.updated_at
  FROM kb_merge_backup.renamed r
  JOIN kb_merge_backup.spaces b ON b.id = r.id
 WHERE r.kind = 'space_key' AND s.id = r.id AND s."key" = r.new_value;

UPDATE principals p SET name = r.old_value, updated_at = b.updated_at
  FROM kb_merge_backup.renamed r
  JOIN kb_merge_backup.principals b ON b.id = r.id
 WHERE r.kind = 'group_name' AND p.id = r.id AND p.name = r.new_value;

-- ============================================================================
-- 8. 個人ワークスペースの is_active を戻す
-- ============================================================================
UPDATE workspaces w SET is_active = b.is_active, updated_at = b.updated_at
  FROM kb_merge_backup.workspaces b
 WHERE w.id = b.id AND (w.is_active IS DISTINCT FROM b.is_active OR w.updated_at IS DISTINCT FROM b.updated_at);

-- ============================================================================
-- 9. 検証 — 控えた行が 1 行残らず元のワークスペースへ戻っていること
-- ============================================================================
DO $verify$
DECLARE
    v        record;
    problems text := '';
BEGIN
    FOR v IN
        SELECT * FROM (VALUES
        ('戻っていない spaces', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.spaces b
              WHERE NOT EXISTS (SELECT 1 FROM spaces x WHERE x.id = b.id AND x.workspace_id = b.workspace_id))),
        ('戻っていない pages', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.pages b
              WHERE NOT EXISTS (SELECT 1 FROM pages x WHERE x.id = b.id AND x.workspace_id = b.workspace_id))),
        ('戻っていない blocks', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.blocks b
              WHERE NOT EXISTS (SELECT 1 FROM blocks x WHERE x.id = b.id AND x.workspace_id = b.workspace_id))),
        ('戻っていない page_paths', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.page_paths b
              WHERE NOT EXISTS (SELECT 1 FROM page_paths x
                                 WHERE x.page_id = b.page_id AND x.ancestor_id = b.ancestor_id
                                   AND x.workspace_id = b.workspace_id))),
        ('戻っていない principals', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.principals b
              WHERE NOT EXISTS (SELECT 1 FROM principals x WHERE x.id = b.id AND x.workspace_id = b.workspace_id))),
        ('戻っていない principal_members', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.principal_members b
              WHERE NOT EXISTS (SELECT 1 FROM principal_members x
                                 WHERE x.group_principal_id = b.group_principal_id
                                   AND x.member_principal_id = b.member_principal_id
                                   AND x.workspace_id = b.workspace_id))),
        ('戻っていない workspace_grants', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.workspace_grants b
              WHERE NOT EXISTS (SELECT 1 FROM workspace_grants x
                                 WHERE x.workspace_id = b.workspace_id AND x.principal_id = b.principal_id))),
        ('戻っていない space_grants', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.space_grants b
              WHERE NOT EXISTS (SELECT 1 FROM space_grants x
                                 WHERE x.workspace_id = b.workspace_id AND x.space_id = b.space_id
                                   AND x.principal_id = b.principal_id AND x."role" = b."role"))),
        ('戻っていない page_restrictions', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.page_restrictions b
              WHERE NOT EXISTS (SELECT 1 FROM page_restrictions x
                                 WHERE x.workspace_id = b.workspace_id AND x.page_id = b.page_id
                                   AND x.principal_id = b.principal_id AND x.capability = b.capability))),
        ('戻っていない page_allow_lists', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.page_allow_lists b
              WHERE NOT EXISTS (SELECT 1 FROM page_allow_lists x
                                 WHERE x.workspace_id = b.workspace_id AND x.page_id = b.page_id
                                   AND x.capability = b.capability))),
        ('戻っていない share_links', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.share_links b
              WHERE NOT EXISTS (SELECT 1 FROM share_links x WHERE x.id = b.id AND x.workspace_id = b.workspace_id))),
        ('戻し損ねた spaces.key', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.renamed r JOIN spaces s ON s.id = r.id
              WHERE r.kind = 'space_key' AND s."key" <> r.old_value)),
        ('戻し損ねたグループ名', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.renamed r JOIN principals p ON p.id = r.id
              WHERE r.kind = 'group_name' AND p.name <> r.old_value)),
        ('統合先に残った移行対象のスペース', 0::bigint,
            (SELECT count(*) FROM spaces x JOIN kb_merge_backup.plan p ON p.target_workspace_id = x.workspace_id
               JOIN kb_merge_backup.spaces b ON b.id = x.id))
        ) AS t(label, expected, actual)
    LOOP
        IF v.expected IS DISTINCT FROM v.actual THEN
            problems := problems || format(E'\n  - %s: 期待 %s / 実際 %s', v.label, v.expected, v.actual);
        END IF;
    END LOOP;
    IF problems <> '' THEN
        RAISE EXCEPTION E'切り戻しの検証に失敗した。トランザクションを巻き戻す:%', problems;
    END IF;
END
$verify$;

-- ============================================================================
-- 10. 控えを片付ける — 移行前とまったく同じ状態（控えも無い）に戻す。
--     こうしておかないと、次に 0024 を流したとき控えに同じ行が二重に積まれ、
--     以後の切り戻しが「どちらの控えが正か」を決められなくなる。
--     いつ戻したかの記録だけは残す。
-- ============================================================================
CREATE TABLE IF NOT EXISTS kb_merge_backup.rollback_log (
    rolled_back_at timestamptz NOT NULL DEFAULT now(),
    run_ids        text        NOT NULL,
    source_count   bigint      NOT NULL
);
INSERT INTO kb_merge_backup.rollback_log (run_ids, source_count)
SELECT string_agg(DISTINCT run_id::text, ','), count(*) FROM kb_merge_backup.plan WHERE merged_at IS NOT NULL;

DELETE FROM kb_merge_backup.share_links       b USING kb_merge_backup.plan p WHERE p.source_workspace_id = b.workspace_id AND p.merged_at IS NOT NULL;
DELETE FROM kb_merge_backup.page_allow_lists  b USING kb_merge_backup.plan p WHERE p.source_workspace_id = b.workspace_id AND p.merged_at IS NOT NULL;
DELETE FROM kb_merge_backup.page_restrictions b USING kb_merge_backup.plan p WHERE p.source_workspace_id = b.workspace_id AND p.merged_at IS NOT NULL;
DELETE FROM kb_merge_backup.space_grants      b USING kb_merge_backup.plan p WHERE p.source_workspace_id = b.workspace_id AND p.merged_at IS NOT NULL;
DELETE FROM kb_merge_backup.workspace_grants  b USING kb_merge_backup.plan p WHERE p.source_workspace_id = b.workspace_id AND p.merged_at IS NOT NULL;
DELETE FROM kb_merge_backup.principal_members b USING kb_merge_backup.plan p WHERE p.source_workspace_id = b.workspace_id AND p.merged_at IS NOT NULL;
DELETE FROM kb_merge_backup.principals        b USING kb_merge_backup.plan p WHERE p.source_workspace_id = b.workspace_id AND p.merged_at IS NOT NULL;
DELETE FROM kb_merge_backup.page_paths        b USING kb_merge_backup.plan p WHERE p.source_workspace_id = b.workspace_id AND p.merged_at IS NOT NULL;
DELETE FROM kb_merge_backup.blocks            b USING kb_merge_backup.plan p WHERE p.source_workspace_id = b.workspace_id AND p.merged_at IS NOT NULL;
DELETE FROM kb_merge_backup.pages             b USING kb_merge_backup.plan p WHERE p.source_workspace_id = b.workspace_id AND p.merged_at IS NOT NULL;
DELETE FROM kb_merge_backup.spaces            b USING kb_merge_backup.plan p WHERE p.source_workspace_id = b.workspace_id AND p.merged_at IS NOT NULL;
DELETE FROM kb_merge_backup.workspaces        b USING kb_merge_backup.plan p WHERE p.source_workspace_id = b.id             AND p.merged_at IS NOT NULL;

DELETE FROM kb_merge_backup.renamed r
 WHERE r.kind = 'space_key'
   AND EXISTS (SELECT 1 FROM spaces s JOIN kb_merge_backup.plan p ON p.source_workspace_id = s.workspace_id
                WHERE s.id = r.id AND p.merged_at IS NOT NULL);
DELETE FROM kb_merge_backup.renamed r
 WHERE r.kind = 'group_name'
   AND EXISTS (SELECT 1 FROM principals g JOIN kb_merge_backup.plan p ON p.source_workspace_id = g.workspace_id
                WHERE g.id = r.id AND p.merged_at IS NOT NULL);

DELETE FROM kb_merge_backup.created_space_grant     WHERE run_id IN (SELECT run_id FROM kb_merge_backup.plan WHERE merged_at IS NOT NULL);
DELETE FROM kb_merge_backup.created_workspace_grant WHERE run_id IN (SELECT run_id FROM kb_merge_backup.plan WHERE merged_at IS NOT NULL);
DELETE FROM kb_merge_backup.created_principal       WHERE run_id IN (SELECT run_id FROM kb_merge_backup.plan WHERE merged_at IS NOT NULL);
DELETE FROM kb_merge_backup.principal_map           WHERE run_id IN (SELECT run_id FROM kb_merge_backup.plan WHERE merged_at IS NOT NULL);
DELETE FROM kb_merge_backup.pre_counts_target       WHERE run_id IN (SELECT run_id FROM kb_merge_backup.plan WHERE merged_at IS NOT NULL);
DELETE FROM kb_merge_backup.pre_counts_global       WHERE run_id IN (SELECT run_id FROM kb_merge_backup.plan WHERE merged_at IS NOT NULL);
DELETE FROM kb_merge_backup.moved_counts            WHERE run_id IN (SELECT run_id FROM kb_merge_backup.plan WHERE merged_at IS NOT NULL);
DELETE FROM kb_merge_backup.plan                    WHERE merged_at IS NOT NULL;

COMMIT;
