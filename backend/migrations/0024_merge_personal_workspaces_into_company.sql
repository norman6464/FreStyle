-- 0024 (Data): 個人ワークスペースの中身を、その人の会社のワークスペースへ統合する。
--
-- 【何が起きているか】
--   会社ごとのワークスペースは起動時のバックフィル（tenant_bridge.go）が自動で作る。
--   一方、実際のノートは各人が UI から作った個人ワークスペースの中にあり、会社の
--   ワークスペースは空のままだった。「同じ会社なら全員が見える」ようにしても、
--   全員が見るのは空の入れ物なので、誰のノートも誰にも見えない。
--
-- 【何をするか】
--   個人ワークスペース配下の全行（spaces / pages / blocks / page_paths / principals /
--   principal_members / workspace_grants / space_grants / page_restrictions /
--   page_allow_lists / share_links）の workspace_id を、その人の会社のワークスペースへ移す。
--   スペースは「会社ワークスペースの中のスペース」として生き残り、visibility='workspace'
--   のスペースは会社の全員（workspace_grants を持つ人）に見える。
--
-- 【会社ワークスペースの決め方（ハードコードしない）】
--   companies.workspace_id から引く。個人ワークスペースの「所属会社」は、その中の
--   principals(kind='user') → users.company_id で決める。**中にいる人が全員同じ 1 社に
--   属しているときだけ**統合する。会社が混ざっている／会社に属さない人がいる
--   ワークスペースは対象外にして NOTICE で報告する（黙って既定の会社へ流し込むと、
--   その会社の人でない人を会社ワークスペースのメンバーにしてしまう）。
--
-- 【なぜ UPDATE の順序ではなく「1 文」なのか（このスキーマ特有の事情）】
--   親子の FK はすべて (workspace_id, ...) を含む複合 FK で、**1 本も DEFERRABLE ではない**
--   （pg_constraint.condeferrable = false）。よって SET CONSTRAINTS ALL DEFERRED は効かない。
--   非 DEFERRABLE の FK は「文の終わり」に検査されるため、
--     - 親を先に動かす → 子がまだ旧 workspace_id を指していて NO ACTION 違反
--       （update or delete on table "spaces" violates fk_pages_space）
--     - 子を先に動かす → 参照先の親がまだ移動していなくて FK 違反
--       （update on table "pages" violates fk_blocks_page … の連鎖）
--   となり、**どちらの順序でも必ず落ちる**（実 PostgreSQL 17.6 で両方試して確認済み）。
--   そこで「移動を全部まとめて 1 文にする」。データ変更 CTE は 1 つの文として実行され、
--   FK トリガは文の終わりにまとめて発火するので、途中の不整合は検査されない。
--   制約の張り替え（DROP/ADD）にも DEFERRABLE 化にも手を出さずに済む唯一の道。
--
-- 【取り返しのつけ方】
--   書き換える前に、対象の全行をそのまま kb_merge_backup スキーマへ退避する。統合先に
--   作った行（principals / workspace_grants / space_grants）と、衝突回避で改名した値も残す。
--   切り戻しは backend/migrations/0024_merge_personal_workspaces_into_company_rollback.sql。
--
-- 【冪等】
--   kb_merge_backup.plan.merged_at が「済み」の印。2 回目は対象 0 件で何も動かず、
--   検証も行わない（何も変えていないので突き合わせる相手がいない）。
--   2 回目以降に新しい個人ワークスペースができていれば、それは新しい実行（run）として
--   統合される。控えと検証は run 単位で持つ。
--
-- 【空になった個人ワークスペースは消さない】
--   workspaces を DELETE すると spaces / principals へ ON DELETE CASCADE が走る。万一
--   「移し損ねた行」が 1 つでも残っていた場合、エラーではなく**沈黙の消失**になる。また
--   users.workspace_id / companies.workspace_id からの FK は NO ACTION なので、指されている
--   行はそもそも消せない。所属の唯一の表現である principals(kind='user') を統合先へ畳んだ
--   時点で、この入れ物は誰の一覧にも出なくなる（ListMemberWorkspaces は principals だけを
--   見る）。行は is_active=false の墓標として残し、本当に消すかは統合が定着してから
--   別途判断する（このファイルの末尾に手順を書いてある）。

BEGIN;

SET LOCAL lock_timeout = '30s';
SET LOCAL statement_timeout = '10min';

-- ============================================================================
-- 0. 前提の点検 — このファイルは「移す表」を名指しで列挙して書いてある。
--    列挙にない workspace_id 列が増えていたら、扱いを決めるまで流させない。
-- ============================================================================
DO $guard$
DECLARE
    unknown text;
BEGIN
    SELECT string_agg(table_name, ', ' ORDER BY table_name) INTO unknown
      FROM information_schema.columns
     WHERE table_schema = current_schema()
       AND column_name = 'workspace_id'
       -- テナント配下の表（この移行が移す）
       AND table_name NOT IN ('spaces','pages','blocks','page_paths','principals',
                              'principal_members','workspace_grants','space_grants',
                              'page_restrictions','page_allow_lists','share_links')
       -- workspaces を「指す」側。テナント配下ではないので移さない
       AND table_name NOT IN ('companies','users');
    IF unknown IS NOT NULL THEN
        RAISE EXCEPTION
            'workspace_id を持つ未知の表がある: %。この移行はテナント配下の表を名指しで移すので、足された表の扱いを決めてから流すこと',
            unknown;
    END IF;
END
$guard$;

-- ============================================================================
-- 1. 退避先と、この実行（run）の番号
--    控えと検証は run 単位で持つ。そうしないと「後日また統合を流す」ときに、
--    前回の控えを基準に突き合わせてしまい、通常の利用で増えたページを差分と誤認する。
-- ============================================================================
CREATE SCHEMA IF NOT EXISTS kb_merge_backup;
COMMENT ON SCHEMA kb_merge_backup IS
    '0024 個人ワークスペース統合の退避先。切り戻し（0024_..._rollback.sql）が使う。統合が定着したら DROP SCHEMA ... CASCADE してよい';

CREATE SEQUENCE IF NOT EXISTS kb_merge_backup.run_seq;
-- この文以降、同じセッション内では currval('kb_merge_backup.run_seq') が今回の run 番号。
SELECT nextval('kb_merge_backup.run_seq');

-- ============================================================================
-- 2. 統合計画 — どの個人ワークスペースを、どの会社ワークスペースへ畳むか
-- ============================================================================
CREATE TABLE IF NOT EXISTS kb_merge_backup.plan (
    source_workspace_id uuid PRIMARY KEY,
    target_workspace_id uuid        NOT NULL,
    company_id          bigint      NOT NULL,
    source_slug         text        NOT NULL,
    source_name         text        NOT NULL,
    source_is_active    boolean,
    run_id              bigint      NOT NULL,
    planned_at          timestamptz NOT NULL DEFAULT now(),
    -- NULL = 未処理。この列が冪等性の印で、以降の全ステップが merged_at IS NULL で絞る。
    merged_at           timestamptz
);

INSERT INTO kb_merge_backup.plan
       (source_workspace_id, target_workspace_id, company_id,
        source_slug, source_name, source_is_active, run_id)
SELECT w.id, c.workspace_id, c.id, w.slug, w.name, w.is_active,
       currval('kb_merge_backup.run_seq')
  FROM workspaces w
  -- そのワークスペースにいる人たちの所属会社。
  --   n_company = 1 … 会社が 1 つに定まる
  --   n_orphan  = 0 … 会社に属さない人が混じっていない
  -- どちらかを満たさないワークスペースは対象外（下の NOTICE で報告する）。
  JOIN LATERAL (
        SELECT count(DISTINCT u.company_id)                        AS n_company,
               count(*) FILTER (WHERE u.company_id IS NULL)        AS n_orphan,
               min(u.company_id)                                   AS company_id
          FROM principals p
          JOIN users u ON u.id = p.user_id
         WHERE p.workspace_id = w.id AND p.kind = 'user'
       ) m ON m.n_company = 1 AND m.n_orphan = 0
  JOIN companies c ON c.id = m.company_id AND c.workspace_id IS NOT NULL
 -- 会社ワークスペース自身は移す側ではない。
 WHERE NOT EXISTS (SELECT 1 FROM companies c2 WHERE c2.workspace_id = w.id)
ON CONFLICT (source_workspace_id) DO NOTHING;

-- 対象外になったワークスペース（中身があるのに統合先が決まらないもの）を報告する。
DO $skipped$
DECLARE
    row_text text;
BEGIN
    SELECT string_agg(format(E'\n  - %s（%s） spaces=%s pages=%s', w.name, w.id,
                             (SELECT count(*) FROM spaces s WHERE s.workspace_id = w.id),
                             (SELECT count(*) FROM pages  p WHERE p.workspace_id = w.id)), '')
      INTO row_text
      FROM workspaces w
     WHERE NOT EXISTS (SELECT 1 FROM companies c  WHERE c.workspace_id = w.id)
       AND NOT EXISTS (SELECT 1 FROM kb_merge_backup.plan p WHERE p.source_workspace_id = w.id)
       AND (EXISTS (SELECT 1 FROM spaces s WHERE s.workspace_id = w.id)
         OR EXISTS (SELECT 1 FROM pages  p WHERE p.workspace_id = w.id));
    IF row_text IS NOT NULL THEN
        RAISE NOTICE
            E'統合先を決められないため対象外にしたワークスペース（中身あり）。会社が混ざっている／会社に属さない人がいる:%',
            row_text;
    END IF;
END
$skipped$;

-- ============================================================================
-- 3. 移行前の状態を退避（書き換えの前に必ず）
--    CREATE TABLE ... AS ... WITH NO DATA で列だけ写す（制約は写らない＝退避先として正しい）。
--    INSERT の対象は「未処理の計画に載っている source の行」だけなので、2 回目の実行では
--    0 件になり、退避が移行後の姿で上書きされることはない。
-- ============================================================================
CREATE TABLE IF NOT EXISTS kb_merge_backup.workspaces        AS SELECT * FROM public.workspaces        WITH NO DATA;
CREATE TABLE IF NOT EXISTS kb_merge_backup.spaces            AS SELECT * FROM public.spaces            WITH NO DATA;
CREATE TABLE IF NOT EXISTS kb_merge_backup.pages             AS SELECT * FROM public.pages             WITH NO DATA;
CREATE TABLE IF NOT EXISTS kb_merge_backup.blocks            AS SELECT * FROM public.blocks            WITH NO DATA;
CREATE TABLE IF NOT EXISTS kb_merge_backup.page_paths        AS SELECT * FROM public.page_paths        WITH NO DATA;
CREATE TABLE IF NOT EXISTS kb_merge_backup.principals        AS SELECT * FROM public.principals        WITH NO DATA;
CREATE TABLE IF NOT EXISTS kb_merge_backup.principal_members AS SELECT * FROM public.principal_members WITH NO DATA;
CREATE TABLE IF NOT EXISTS kb_merge_backup.workspace_grants  AS SELECT * FROM public.workspace_grants  WITH NO DATA;
CREATE TABLE IF NOT EXISTS kb_merge_backup.space_grants      AS SELECT * FROM public.space_grants      WITH NO DATA;
CREATE TABLE IF NOT EXISTS kb_merge_backup.page_restrictions AS SELECT * FROM public.page_restrictions WITH NO DATA;
CREATE TABLE IF NOT EXISTS kb_merge_backup.page_allow_lists  AS SELECT * FROM public.page_allow_lists  WITH NO DATA;
CREATE TABLE IF NOT EXISTS kb_merge_backup.share_links       AS SELECT * FROM public.share_links       WITH NO DATA;

INSERT INTO kb_merge_backup.workspaces
SELECT w.* FROM public.workspaces w
  JOIN kb_merge_backup.plan p ON p.source_workspace_id = w.id AND p.merged_at IS NULL;
INSERT INTO kb_merge_backup.spaces
SELECT x.* FROM public.spaces x
  JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL;
INSERT INTO kb_merge_backup.pages
SELECT x.* FROM public.pages x
  JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL;
INSERT INTO kb_merge_backup.blocks
SELECT x.* FROM public.blocks x
  JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL;
INSERT INTO kb_merge_backup.page_paths
SELECT x.* FROM public.page_paths x
  JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL;
INSERT INTO kb_merge_backup.principals
SELECT x.* FROM public.principals x
  JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL;
INSERT INTO kb_merge_backup.principal_members
SELECT x.* FROM public.principal_members x
  JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL;
INSERT INTO kb_merge_backup.workspace_grants
SELECT x.* FROM public.workspace_grants x
  JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL;
INSERT INTO kb_merge_backup.space_grants
SELECT x.* FROM public.space_grants x
  JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL;
INSERT INTO kb_merge_backup.page_restrictions
SELECT x.* FROM public.page_restrictions x
  JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL;
INSERT INTO kb_merge_backup.page_allow_lists
SELECT x.* FROM public.page_allow_lists x
  JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL;
INSERT INTO kb_merge_backup.share_links
SELECT x.* FROM public.share_links x
  JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL;

-- 件数の控え。移行後に「1 行も増えず減らず、移っただけ」を突き合わせるために使う。
-- page_snapshots は blocks から作り直せる派生データなので中身は退避せず、件数だけ見る
-- （CASCADE で巻き添えに消えていないかは、この件数で分かる）。
CREATE TABLE IF NOT EXISTS kb_merge_backup.pre_counts_global (
    run_id     bigint NOT NULL,
    table_name text   NOT NULL,
    n          bigint NOT NULL,
    taken_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (run_id, table_name)
);
CREATE TABLE IF NOT EXISTS kb_merge_backup.pre_counts_target (
    run_id       bigint NOT NULL,
    workspace_id uuid   NOT NULL,
    table_name   text   NOT NULL,
    n            bigint NOT NULL,
    taken_at     timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (run_id, workspace_id, table_name)
);
CREATE TABLE IF NOT EXISTS kb_merge_backup.moved_counts (
    run_id     bigint NOT NULL,
    table_name text   NOT NULL,
    n          bigint NOT NULL,
    moved_at   timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (run_id, table_name)
);

INSERT INTO kb_merge_backup.pre_counts_global (run_id, table_name, n)
SELECT currval('kb_merge_backup.run_seq'), v.table_name, v.n
  FROM (VALUES
        ('spaces',            (SELECT count(*) FROM spaces)),
        ('pages',             (SELECT count(*) FROM pages)),
        ('blocks',            (SELECT count(*) FROM blocks)),
        ('page_paths',        (SELECT count(*) FROM page_paths)),
        ('page_snapshots',    (SELECT count(*) FROM page_snapshots)),
        ('principals',        (SELECT count(*) FROM principals)),
        ('principal_members', (SELECT count(*) FROM principal_members)),
        ('workspace_grants',  (SELECT count(*) FROM workspace_grants)),
        ('space_grants',      (SELECT count(*) FROM space_grants)),
        ('page_restrictions', (SELECT count(*) FROM page_restrictions)),
        ('page_allow_lists',  (SELECT count(*) FROM page_allow_lists)),
        ('share_links',       (SELECT count(*) FROM share_links))
       ) AS v(table_name, n)
 -- 未処理の計画が無いなら控えも取らない（何もしない実行では基準も作らない）
 WHERE EXISTS (SELECT 1 FROM kb_merge_backup.plan WHERE merged_at IS NULL)
ON CONFLICT (run_id, table_name) DO NOTHING;

INSERT INTO kb_merge_backup.pre_counts_target (run_id, workspace_id, table_name, n)
SELECT currval('kb_merge_backup.run_seq'), t.ws, v.table_name, v.n
  FROM (SELECT DISTINCT target_workspace_id AS ws
          FROM kb_merge_backup.plan WHERE merged_at IS NULL) t
 CROSS JOIN LATERAL (VALUES
        ('spaces',            (SELECT count(*) FROM spaces            x WHERE x.workspace_id = t.ws)),
        ('pages',             (SELECT count(*) FROM pages             x WHERE x.workspace_id = t.ws)),
        ('blocks',            (SELECT count(*) FROM blocks            x WHERE x.workspace_id = t.ws)),
        ('page_paths',        (SELECT count(*) FROM page_paths        x WHERE x.workspace_id = t.ws)),
        ('principals',        (SELECT count(*) FROM principals        x WHERE x.workspace_id = t.ws)),
        ('principal_members', (SELECT count(*) FROM principal_members x WHERE x.workspace_id = t.ws)),
        ('workspace_grants',  (SELECT count(*) FROM workspace_grants  x WHERE x.workspace_id = t.ws)),
        ('space_grants',      (SELECT count(*) FROM space_grants      x WHERE x.workspace_id = t.ws)),
        ('page_restrictions', (SELECT count(*) FROM page_restrictions x WHERE x.workspace_id = t.ws)),
        ('page_allow_lists',  (SELECT count(*) FROM page_allow_lists  x WHERE x.workspace_id = t.ws)),
        ('share_links',       (SELECT count(*) FROM share_links       x WHERE x.workspace_id = t.ws))
       ) AS v(table_name, n)
ON CONFLICT (run_id, workspace_id, table_name) DO NOTHING;

-- ============================================================================
-- 4. 衝突の解消（統合先へ入れる前に、一意制約に触れる値を直しておく）
--
--    一意「索引」は制約と違って遅延できず、1 行ごとに即座に検査される。したがって
--    6 の 1 文に入る前に、衝突が起きない状態にしておく必要がある。
-- ============================================================================
CREATE TABLE IF NOT EXISTS kb_merge_backup.renamed (
    kind      text   NOT NULL,   -- 'space_key' | 'group_name'
    id        uuid   NOT NULL,
    old_value text   NOT NULL,
    new_value text   NOT NULL,
    run_id    bigint NOT NULL,
    PRIMARY KEY (kind, id)
);

-- 4a. spaces の key は UNIQUE (workspace_id, "key")。統合先に同じ key があるもの／
--     source 同士で同じ key のもの（最初の 1 つ以外）を改名する。新しい値は id から
--     決まる（md5 の先頭 12 桁）ので、実行のたびに変わらず、互いにも衝突しない。
WITH cand AS (
    SELECT s.id, s."key", p.target_workspace_id AS tgt,
           row_number() OVER (PARTITION BY p.target_workspace_id, s."key" ORDER BY s.id) AS rn
      FROM spaces s
      JOIN kb_merge_backup.plan p
        ON p.source_workspace_id = s.workspace_id AND p.merged_at IS NULL
), need AS (
    SELECT c.id, c."key" AS old_key,
           left(c."key", 50) || '-' || substr(md5(c.id::text), 1, 12) AS new_key
      FROM cand c
     WHERE c.rn > 1
        OR EXISTS (SELECT 1 FROM spaces t WHERE t.workspace_id = c.tgt AND t."key" = c."key")
), upd AS (
    UPDATE spaces s
       SET "key" = n.new_key, updated_at = now()
      FROM need n
     WHERE s.id = n.id
    RETURNING s.id, n.old_key, n.new_key
)
INSERT INTO kb_merge_backup.renamed (kind, id, old_value, new_value, run_id)
SELECT 'space_key', id, old_key, new_key, currval('kb_merge_backup.run_seq') FROM upd
ON CONFLICT (kind, id) DO NOTHING;

-- 4b. グループ名は UNIQUE (workspace_id, name) WHERE kind='group'。同じ考え方で改名する。
WITH cand AS (
    SELECT g.id, g.name, p.target_workspace_id AS tgt,
           row_number() OVER (PARTITION BY p.target_workspace_id, g.name ORDER BY g.id) AS rn
      FROM principals g
      JOIN kb_merge_backup.plan p
        ON p.source_workspace_id = g.workspace_id AND p.merged_at IS NULL
     WHERE g.kind = 'group'
), need AS (
    SELECT c.id, c.name AS old_name,
           left(c.name, 180) || ' (' || substr(md5(c.id::text), 1, 8) || ')' AS new_name
      FROM cand c
     WHERE c.rn > 1
        OR EXISTS (SELECT 1 FROM principals t
                    WHERE t.workspace_id = c.tgt AND t.kind = 'group' AND t.name = c.name)
), upd AS (
    UPDATE principals g
       SET name = n.new_name, updated_at = now()
      FROM need n
     WHERE g.id = n.id
    RETURNING g.id, n.old_name, n.new_name
)
INSERT INTO kb_merge_backup.renamed (kind, id, old_value, new_value, run_id)
SELECT 'group_name', id, old_name, new_name, currval('kb_merge_backup.run_seq') FROM upd
ON CONFLICT (kind, id) DO NOTHING;

-- 改名しきれたか（統合先 ∪ source で key / グループ名が一意か）をここで確かめて止める。
DO $collision$
DECLARE
    dup_keys   bigint;
    dup_groups bigint;
BEGIN
    SELECT count(*) INTO dup_keys FROM (
        SELECT 1 FROM (
            SELECT p.target_workspace_id AS ws, s."key" AS v
              FROM spaces s JOIN kb_merge_backup.plan p
                ON p.source_workspace_id = s.workspace_id AND p.merged_at IS NULL
             UNION ALL
            SELECT s.workspace_id, s."key" FROM spaces s
             WHERE s.workspace_id IN (SELECT target_workspace_id FROM kb_merge_backup.plan
                                       WHERE merged_at IS NULL)) t
         GROUP BY t.ws, t.v HAVING count(*) > 1) d;
    SELECT count(*) INTO dup_groups FROM (
        SELECT 1 FROM (
            SELECT p.target_workspace_id AS ws, g.name AS v
              FROM principals g JOIN kb_merge_backup.plan p
                ON p.source_workspace_id = g.workspace_id AND p.merged_at IS NULL
             WHERE g.kind = 'group'
             UNION ALL
            SELECT g.workspace_id, g.name FROM principals g
             WHERE g.kind = 'group'
               AND g.workspace_id IN (SELECT target_workspace_id FROM kb_merge_backup.plan
                                       WHERE merged_at IS NULL)) t
         GROUP BY t.ws, t.v HAVING count(*) > 1) d;
    IF dup_keys > 0 OR dup_groups > 0 THEN
        RAISE EXCEPTION '改名しても衝突が残っている（spaces.key: % 組 / グループ名: % 組）。手で解消してから流すこと',
            dup_keys, dup_groups;
    END IF;
END
$collision$;

-- ============================================================================
-- 5. 統合先に「その人の主体」と既定の役割を用意する
--
--    principals(kind='user') はワークスペース所属の唯一の表現で、
--    UNIQUE (workspace_id, user_id) WHERE kind='user' がある。同じ人が個人ワークスペースを
--    3 つ持っていても、統合先に置ける主体は 1 つ。よって source 側の user 主体は
--    「移す」のではなく「統合先の主体へ参照を付け替えてから捨てる」。
--
--    役割（workspace_grants）を必ず一緒に作るのが要点。JoinCompanyWorkspaceUseCase は
--    「主体を新しく作ったときだけ」既定の役割を与える。ここで主体だけ作って役割を作らないと、
--    以後その人は「所属しているのに何の役割も無い＝何も見えない」まま固定される。
-- ============================================================================
CREATE TABLE IF NOT EXISTS kb_merge_backup.created_principal (
    id           uuid PRIMARY KEY,
    workspace_id uuid   NOT NULL,
    user_id      bigint NOT NULL,
    run_id       bigint NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE IF NOT EXISTS kb_merge_backup.created_workspace_grant (
    workspace_id uuid   NOT NULL,
    principal_id uuid   NOT NULL,
    "role"       text   NOT NULL,
    run_id       bigint NOT NULL,
    PRIMARY KEY (workspace_id, principal_id)
);
CREATE TABLE IF NOT EXISTS kb_merge_backup.created_space_grant (
    workspace_id uuid   NOT NULL,
    space_id     uuid   NOT NULL,
    principal_id uuid   NOT NULL,
    "role"       text   NOT NULL,
    run_id       bigint NOT NULL,
    PRIMARY KEY (workspace_id, space_id, principal_id)
);
-- source の主体 → 統合先の主体。6 の 1 文と切り戻しの両方がこの対応表を使う。
CREATE TABLE IF NOT EXISTS kb_merge_backup.principal_map (
    source_principal_id uuid PRIMARY KEY,
    source_workspace_id uuid   NOT NULL,
    target_workspace_id uuid   NOT NULL,
    target_principal_id uuid   NOT NULL,
    user_id             bigint NOT NULL,
    run_id              bigint NOT NULL
);

-- 5a. 足りない主体を作る。id はアプリと同じ UUIDv7 相当（先頭 48bit = ミリ秒）で採番する
--     — アプリが書く行だけが時系列に並び、移行が書いた行だけ並ばない、という差を残さないため。
--     組み立て: ミリ秒(12桁) + バージョン'7' + 乱数(3) + バリアント(8..b) + 乱数(15) = 32 桁。
WITH want AS (
    SELECT DISTINCT p.target_workspace_id AS ws, sp.user_id
      FROM kb_merge_backup.plan p
      JOIN principals sp ON sp.workspace_id = p.source_workspace_id AND sp.kind = 'user'
     WHERE p.merged_at IS NULL
), ins AS (
    INSERT INTO principals (id, workspace_id, kind, user_id)
    SELECT (lpad(to_hex((extract(epoch FROM clock_timestamp()) * 1000)::bigint), 12, '0')
            || '7' || substr(md5(random()::text || clock_timestamp()::text), 1, 3)
            || to_hex(8 + (random() * 3)::int)
            || substr(md5(random()::text || gen_random_uuid()::text), 1, 15))::uuid,
           w.ws, 'user', w.user_id
      FROM want w
     WHERE NOT EXISTS (SELECT 1 FROM principals t
                        WHERE t.workspace_id = w.ws AND t.kind = 'user' AND t.user_id = w.user_id)
    RETURNING id, workspace_id, user_id
)
INSERT INTO kb_merge_backup.created_principal (id, workspace_id, user_id, run_id)
SELECT id, workspace_id, user_id, currval('kb_merge_backup.run_seq') FROM ins;

-- 5b. 新しく作った主体に既定の役割（editor）を与える。JoinCompanyWorkspaceUseCase と同じ値。
WITH ins AS (
    INSERT INTO workspace_grants (workspace_id, principal_id, "role")
    SELECT c.workspace_id, c.id, 'editor'
      FROM kb_merge_backup.created_principal c
     WHERE c.run_id = currval('kb_merge_backup.run_seq')
       AND NOT EXISTS (SELECT 1 FROM workspace_grants g
                        WHERE g.workspace_id = c.workspace_id AND g.principal_id = c.id)
    RETURNING workspace_id, principal_id, "role"
)
INSERT INTO kb_merge_backup.created_workspace_grant (workspace_id, principal_id, "role", run_id)
SELECT workspace_id, principal_id, "role", currval('kb_merge_backup.run_seq') FROM ins
ON CONFLICT (workspace_id, principal_id) DO NOTHING;

-- 5c. 対応表を作る。
INSERT INTO kb_merge_backup.principal_map
       (source_principal_id, source_workspace_id, target_workspace_id,
        target_principal_id, user_id, run_id)
SELECT sp.id, sp.workspace_id, p.target_workspace_id, tp.id, sp.user_id,
       currval('kb_merge_backup.run_seq')
  FROM kb_merge_backup.plan p
  JOIN principals sp ON sp.workspace_id = p.source_workspace_id AND sp.kind = 'user'
  JOIN principals tp ON tp.workspace_id = p.target_workspace_id AND tp.kind = 'user'
                    AND tp.user_id = sp.user_id
 WHERE p.merged_at IS NULL
ON CONFLICT (source_principal_id) DO NOTHING;

-- 写像に漏れがあると 6 の 1 文が「消した主体を指す行」を残して FK で落ちる。先に止める。
DO $mapping$
DECLARE
    missing bigint;
BEGIN
    SELECT count(*) INTO missing
      FROM principals sp
      JOIN kb_merge_backup.plan p
        ON p.source_workspace_id = sp.workspace_id AND p.merged_at IS NULL
     WHERE sp.kind = 'user'
       AND NOT EXISTS (SELECT 1 FROM kb_merge_backup.principal_map m
                        WHERE m.source_principal_id = sp.id);
    IF missing > 0 THEN
        RAISE EXCEPTION '統合先の主体に写像できない source の主体が % 件ある', missing;
    END IF;
END
$mapping$;

-- ============================================================================
-- 6. 本体 — テナントの移動。**必ず 1 文**（理由はファイル冒頭）。
--
--    ここに含まれるもの:
--      - workspace_id の書き換え（spaces / pages / blocks / page_paths /
--        page_allow_lists / page_restrictions / space_grants / share_links /
--        principal_members / principals(user 以外)）
--      - principal_id の付け替え（space_grants / page_restrictions / principal_members）
--      - source の user 主体の削除（参照はすべて同じ文の中で付け替わっている）
--      - source の workspace_grants の削除（次の 7 でスペース単位の役割へ畳み直す）
--
--    updated_at は動かさない。これは「入れ物を移す」移行であって、利用者から見たページの
--    更新ではない（tenant_bridge の写しと同じ判断）。改名した spaces."key" /
--    principals.name だけは値そのものが変わるので 4 で動かしてある。
-- ============================================================================
WITH plan AS (
    SELECT source_workspace_id AS src, target_workspace_id AS tgt
      FROM kb_merge_backup.plan WHERE merged_at IS NULL
),
mv_spaces AS (
    UPDATE spaces x SET workspace_id = p.tgt
      FROM plan p WHERE x.workspace_id = p.src RETURNING 1
),
mv_pages AS (
    UPDATE pages x SET workspace_id = p.tgt
      FROM plan p WHERE x.workspace_id = p.src RETURNING 1
),
mv_blocks AS (
    UPDATE blocks x SET workspace_id = p.tgt
      FROM plan p WHERE x.workspace_id = p.src RETURNING 1
),
mv_page_paths AS (
    UPDATE page_paths x SET workspace_id = p.tgt
      FROM plan p WHERE x.workspace_id = p.src RETURNING 1
),
mv_page_allow_lists AS (
    UPDATE page_allow_lists x SET workspace_id = p.tgt
      FROM plan p WHERE x.workspace_id = p.src RETURNING 1
),
-- user 以外（group / space_all / share_link）の主体はそのまま統合先へ移す。
-- 統合先に同じ人の主体が居る、という衝突が起きるのは kind='user' だけ。
mv_principals AS (
    UPDATE principals x SET workspace_id = p.tgt
      FROM plan p WHERE x.workspace_id = p.src AND x.kind <> 'user' RETURNING 1
),
mv_page_restrictions AS (
    UPDATE page_restrictions x
       SET workspace_id = p.tgt,
           principal_id = COALESCE((SELECT m.target_principal_id FROM kb_merge_backup.principal_map m
                                     WHERE m.source_principal_id = x.principal_id), x.principal_id)
      FROM plan p WHERE x.workspace_id = p.src RETURNING 1
),
mv_space_grants AS (
    UPDATE space_grants x
       SET workspace_id = p.tgt,
           principal_id = COALESCE((SELECT m.target_principal_id FROM kb_merge_backup.principal_map m
                                     WHERE m.source_principal_id = x.principal_id), x.principal_id)
      FROM plan p WHERE x.workspace_id = p.src RETURNING 1
),
mv_principal_members AS (
    UPDATE principal_members x
       SET workspace_id        = p.tgt,
           member_principal_id = COALESCE((SELECT m.target_principal_id FROM kb_merge_backup.principal_map m
                                            WHERE m.source_principal_id = x.member_principal_id),
                                          x.member_principal_id)
      FROM plan p WHERE x.workspace_id = p.src RETURNING 1
),
mv_share_links AS (
    UPDATE share_links x SET workspace_id = p.tgt
      FROM plan p WHERE x.workspace_id = p.src RETURNING 1
),
-- ワークスペース全体の役割は移さない。移すと「自分の個人ワークスペースの admin」が
-- そのまま「会社ワークスペースの admin」になり、全社のノートの権限を触れてしまう。
del_workspace_grants AS (
    DELETE FROM workspace_grants x USING plan p WHERE x.workspace_id = p.src RETURNING 1
),
-- source の user 主体を捨てる。参照はすべて上の CTE で統合先の主体へ付け替わっているので、
-- ON DELETE CASCADE は 1 行も道連れにしない（文末に走る時点で、もう誰も指していない）。
del_user_principals AS (
    DELETE FROM principals x USING plan p
     WHERE x.workspace_id = p.src AND x.kind = 'user' RETURNING 1
)
INSERT INTO kb_merge_backup.moved_counts (run_id, table_name, n)
SELECT currval('kb_merge_backup.run_seq'), s.table_name, s.n FROM (
          SELECT 'spaces' AS table_name, count(*) AS n FROM mv_spaces
UNION ALL SELECT 'pages', count(*) FROM mv_pages
UNION ALL SELECT 'blocks', count(*) FROM mv_blocks
UNION ALL SELECT 'page_paths', count(*) FROM mv_page_paths
UNION ALL SELECT 'page_allow_lists', count(*) FROM mv_page_allow_lists
UNION ALL SELECT 'principals', count(*) FROM mv_principals
UNION ALL SELECT 'page_restrictions', count(*) FROM mv_page_restrictions
UNION ALL SELECT 'space_grants', count(*) FROM mv_space_grants
UNION ALL SELECT 'principal_members', count(*) FROM mv_principal_members
UNION ALL SELECT 'share_links', count(*) FROM mv_share_links
UNION ALL SELECT 'workspace_grants(削除)', count(*) FROM del_workspace_grants
UNION ALL SELECT 'principals(user・削除)', count(*) FROM del_user_principals
) s
-- 対象 0 件の実行では控えも書かない（何もしなかった実行の痕跡を残さない）
WHERE EXISTS (SELECT 1 FROM kb_merge_backup.plan WHERE merged_at IS NULL)
ON CONFLICT (run_id, table_name) DO NOTHING;

-- ============================================================================
-- 7. 消した「ワークスペース全体の役割」を、そのワークスペースが持っていたスペースの
--    役割として置き直す。個人ワークスペースの admin は、統合後は「自分が作ったスペースの
--    admin」になる（会社全体の admin にはしない）。既に役割があるスペースには触らない。
-- ============================================================================
WITH ins AS (
    INSERT INTO space_grants (workspace_id, space_id, principal_id, "role")
    SELECT p.target_workspace_id, bs.id,
           COALESCE(m.target_principal_id, g.principal_id), g."role"
      FROM kb_merge_backup.workspace_grants g
      JOIN kb_merge_backup.plan p
        ON p.source_workspace_id = g.workspace_id AND p.merged_at IS NULL
      JOIN kb_merge_backup.spaces bs ON bs.workspace_id = g.workspace_id
      LEFT JOIN kb_merge_backup.principal_map m ON m.source_principal_id = g.principal_id
    ON CONFLICT (workspace_id, space_id, principal_id) DO NOTHING
    RETURNING workspace_id, space_id, principal_id, "role"
)
INSERT INTO kb_merge_backup.created_space_grant (workspace_id, space_id, principal_id, "role", run_id)
SELECT workspace_id, space_id, principal_id, "role", currval('kb_merge_backup.run_seq') FROM ins
ON CONFLICT (workspace_id, space_id, principal_id) DO NOTHING;

-- ============================================================================
-- 8. 空になった個人ワークスペースの後始末（消さない。理由はファイル冒頭）
-- ============================================================================
UPDATE workspaces w
   SET is_active = false, updated_at = now()
  FROM kb_merge_backup.plan p
 WHERE p.source_workspace_id = w.id
   AND p.merged_at IS NULL
   AND w.is_active IS DISTINCT FROM false;

-- ============================================================================
-- 9. 検証 — 期待と 1 つでも違えば例外で止める（トランザクションごと巻き戻る）
-- ============================================================================
DO $verify$
DECLARE
    v        record;
    run      bigint := currval('kb_merge_backup.run_seq');
    problems text := '';
BEGIN
    IF NOT EXISTS (SELECT 1 FROM kb_merge_backup.plan WHERE merged_at IS NULL) THEN
        -- 何も動かしていない実行。突き合わせる相手（この run の控え）が無いので検証もしない。
        RETURN;
    END IF;

    FOR v IN
        SELECT * FROM (VALUES
        -- (1) source 側に取り残しが無いこと
        ('取り残し spaces', 0::bigint,
            (SELECT count(*) FROM spaces x JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL)),
        ('取り残し pages', 0::bigint,
            (SELECT count(*) FROM pages x JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL)),
        ('取り残し blocks', 0::bigint,
            (SELECT count(*) FROM blocks x JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL)),
        ('取り残し page_paths', 0::bigint,
            (SELECT count(*) FROM page_paths x JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL)),
        ('取り残し principals', 0::bigint,
            (SELECT count(*) FROM principals x JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL)),
        ('取り残し principal_members', 0::bigint,
            (SELECT count(*) FROM principal_members x JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL)),
        ('取り残し workspace_grants', 0::bigint,
            (SELECT count(*) FROM workspace_grants x JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL)),
        ('取り残し space_grants', 0::bigint,
            (SELECT count(*) FROM space_grants x JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL)),
        ('取り残し page_restrictions', 0::bigint,
            (SELECT count(*) FROM page_restrictions x JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL)),
        ('取り残し page_allow_lists', 0::bigint,
            (SELECT count(*) FROM page_allow_lists x JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL)),
        ('取り残し share_links', 0::bigint,
            (SELECT count(*) FROM share_links x JOIN kb_merge_backup.plan p ON p.source_workspace_id = x.workspace_id AND p.merged_at IS NULL)),

        -- (2) 退避した行が 1 行残らず統合先に居ること（id で突き合わせる。件数だけの一致では
        --     「別のワークスペースへ移した」「中身が変わった」を見逃す）
        ('統合先に居ない spaces', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.spaces b
               JOIN kb_merge_backup.plan p ON p.source_workspace_id = b.workspace_id AND p.merged_at IS NULL
              WHERE NOT EXISTS (SELECT 1 FROM spaces x
                                 WHERE x.id = b.id AND x.workspace_id = p.target_workspace_id
                                   AND x.name = b.name AND x.visibility = b.visibility))),
        ('統合先に居ない pages', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.pages b
               JOIN kb_merge_backup.plan p ON p.source_workspace_id = b.workspace_id AND p.merged_at IS NULL
              WHERE NOT EXISTS (SELECT 1 FROM pages x
                                 WHERE x.id = b.id AND x.workspace_id = p.target_workspace_id
                                   AND x.space_id = b.space_id
                                   AND x.parent_id IS NOT DISTINCT FROM b.parent_id
                                   AND x.title = b.title AND x."position" = b."position"
                                   AND x.archived_at IS NOT DISTINCT FROM b.archived_at))),
        ('統合先に居ない blocks', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.blocks b
               JOIN kb_merge_backup.plan p ON p.source_workspace_id = b.workspace_id AND p.merged_at IS NULL
              WHERE NOT EXISTS (SELECT 1 FROM blocks x
                                 WHERE x.id = b.id AND x.workspace_id = p.target_workspace_id
                                   AND x.page_id = b.page_id
                                   AND x.parent_id IS NOT DISTINCT FROM b.parent_id
                                   AND x.type = b.type AND x.attrs = b.attrs
                                   AND x.inline IS NOT DISTINCT FROM b.inline))),
        ('統合先に居ない page_paths', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.page_paths b
               JOIN kb_merge_backup.plan p ON p.source_workspace_id = b.workspace_id AND p.merged_at IS NULL
              WHERE NOT EXISTS (SELECT 1 FROM page_paths x
                                 WHERE x.page_id = b.page_id AND x.ancestor_id = b.ancestor_id
                                   AND x.workspace_id = p.target_workspace_id AND x.depth = b.depth))),
        ('統合先に居ない principals(user 以外)', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.principals b
               JOIN kb_merge_backup.plan p ON p.source_workspace_id = b.workspace_id AND p.merged_at IS NULL
              WHERE b.kind <> 'user'
                AND NOT EXISTS (SELECT 1 FROM principals x
                                 WHERE x.id = b.id AND x.workspace_id = p.target_workspace_id
                                   AND x.kind = b.kind
                                   AND x.space_id IS NOT DISTINCT FROM b.space_id
                                   AND x.page_id  IS NOT DISTINCT FROM b.page_id))),
        ('捨てそこねた principals(user)', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.principals b
               JOIN kb_merge_backup.plan p ON p.source_workspace_id = b.workspace_id AND p.merged_at IS NULL
              WHERE b.kind = 'user' AND EXISTS (SELECT 1 FROM principals x WHERE x.id = b.id))),
        ('統合先に居ない page_restrictions', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.page_restrictions b
               JOIN kb_merge_backup.plan p ON p.source_workspace_id = b.workspace_id AND p.merged_at IS NULL
               LEFT JOIN kb_merge_backup.principal_map m ON m.source_principal_id = b.principal_id
              WHERE NOT EXISTS (SELECT 1 FROM page_restrictions x
                                 WHERE x.workspace_id = p.target_workspace_id AND x.page_id = b.page_id
                                   AND x.principal_id = COALESCE(m.target_principal_id, b.principal_id)
                                   AND x.capability = b.capability AND x.mode = b.mode))),
        ('統合先に居ない page_allow_lists', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.page_allow_lists b
               JOIN kb_merge_backup.plan p ON p.source_workspace_id = b.workspace_id AND p.merged_at IS NULL
              WHERE NOT EXISTS (SELECT 1 FROM page_allow_lists x
                                 WHERE x.workspace_id = p.target_workspace_id AND x.page_id = b.page_id
                                   AND x.capability = b.capability))),
        ('統合先に居ない space_grants', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.space_grants b
               JOIN kb_merge_backup.plan p ON p.source_workspace_id = b.workspace_id AND p.merged_at IS NULL
               LEFT JOIN kb_merge_backup.principal_map m ON m.source_principal_id = b.principal_id
              WHERE NOT EXISTS (SELECT 1 FROM space_grants x
                                 WHERE x.workspace_id = p.target_workspace_id AND x.space_id = b.space_id
                                   AND x.principal_id = COALESCE(m.target_principal_id, b.principal_id)
                                   AND x."role" = b."role"))),
        ('統合先に居ない principal_members', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.principal_members b
               JOIN kb_merge_backup.plan p ON p.source_workspace_id = b.workspace_id AND p.merged_at IS NULL
               LEFT JOIN kb_merge_backup.principal_map m ON m.source_principal_id = b.member_principal_id
              WHERE NOT EXISTS (SELECT 1 FROM principal_members x
                                 WHERE x.workspace_id = p.target_workspace_id
                                   AND x.group_principal_id = b.group_principal_id
                                   AND x.member_principal_id = COALESCE(m.target_principal_id, b.member_principal_id)))),
        ('統合先に居ない share_links', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.share_links b
               JOIN kb_merge_backup.plan p ON p.source_workspace_id = b.workspace_id AND p.merged_at IS NULL
              WHERE NOT EXISTS (SELECT 1 FROM share_links x
                                 WHERE x.id = b.id AND x.workspace_id = p.target_workspace_id
                                   AND x.page_id = b.page_id AND x.principal_id = b.principal_id
                                   AND x.token_hash = b.token_hash))),

        -- (3) 全体の件数が「移っただけ」であること。CASCADE で巻き添えに消えていれば必ずここで落ちる。
        ('全体件数 spaces', (SELECT n FROM kb_merge_backup.pre_counts_global WHERE run_id=run AND table_name='spaces'),
            (SELECT count(*) FROM spaces)),
        ('全体件数 pages', (SELECT n FROM kb_merge_backup.pre_counts_global WHERE run_id=run AND table_name='pages'),
            (SELECT count(*) FROM pages)),
        ('全体件数 blocks', (SELECT n FROM kb_merge_backup.pre_counts_global WHERE run_id=run AND table_name='blocks'),
            (SELECT count(*) FROM blocks)),
        ('全体件数 page_paths', (SELECT n FROM kb_merge_backup.pre_counts_global WHERE run_id=run AND table_name='page_paths'),
            (SELECT count(*) FROM page_paths)),
        ('全体件数 page_snapshots', (SELECT n FROM kb_merge_backup.pre_counts_global WHERE run_id=run AND table_name='page_snapshots'),
            (SELECT count(*) FROM page_snapshots)),
        ('全体件数 page_restrictions', (SELECT n FROM kb_merge_backup.pre_counts_global WHERE run_id=run AND table_name='page_restrictions'),
            (SELECT count(*) FROM page_restrictions)),
        ('全体件数 page_allow_lists', (SELECT n FROM kb_merge_backup.pre_counts_global WHERE run_id=run AND table_name='page_allow_lists'),
            (SELECT count(*) FROM page_allow_lists)),
        ('全体件数 share_links', (SELECT n FROM kb_merge_backup.pre_counts_global WHERE run_id=run AND table_name='share_links'),
            (SELECT count(*) FROM share_links)),
        ('全体件数 principal_members', (SELECT n FROM kb_merge_backup.pre_counts_global WHERE run_id=run AND table_name='principal_members'),
            (SELECT count(*) FROM principal_members)),
        -- principals は「作った分 − 捨てた分」だけ動く
        ('全体件数 principals',
            (SELECT n FROM kb_merge_backup.pre_counts_global WHERE run_id=run AND table_name='principals')
              + (SELECT count(*) FROM kb_merge_backup.created_principal WHERE run_id=run)
              - (SELECT count(*) FROM kb_merge_backup.principals b
                   JOIN kb_merge_backup.plan p ON p.source_workspace_id=b.workspace_id AND p.merged_at IS NULL
                  WHERE b.kind='user'),
            (SELECT count(*) FROM principals)),
        -- workspace_grants は「作った分 − source から消した分」
        ('全体件数 workspace_grants',
            (SELECT n FROM kb_merge_backup.pre_counts_global WHERE run_id=run AND table_name='workspace_grants')
              + (SELECT count(*) FROM kb_merge_backup.created_workspace_grant WHERE run_id=run)
              - (SELECT count(*) FROM kb_merge_backup.workspace_grants b
                   JOIN kb_merge_backup.plan p ON p.source_workspace_id=b.workspace_id AND p.merged_at IS NULL),
            (SELECT count(*) FROM workspace_grants)),
        -- space_grants は「7 で置き直した分」だけ増える
        ('全体件数 space_grants',
            (SELECT n FROM kb_merge_backup.pre_counts_global WHERE run_id=run AND table_name='space_grants')
              + (SELECT count(*) FROM kb_merge_backup.created_space_grant WHERE run_id=run),
            (SELECT count(*) FROM space_grants)),

        -- (4) 統合先の件数 = 移行前の件数 + 移した件数（統合先ごとに突き合わせ、ずれた統合先を数える）
        ('統合先の spaces/pages 件数が合わない統合先', 0::bigint,
            (SELECT count(*) FROM (
                SELECT t.ws
                  FROM (SELECT DISTINCT target_workspace_id AS ws FROM kb_merge_backup.plan WHERE merged_at IS NULL) t
                 WHERE (SELECT count(*) FROM pages x WHERE x.workspace_id = t.ws)
                       <> COALESCE((SELECT n FROM kb_merge_backup.pre_counts_target
                                     WHERE run_id=run AND workspace_id = t.ws AND table_name = 'pages'), 0)
                          + (SELECT count(*) FROM kb_merge_backup.pages b
                               JOIN kb_merge_backup.plan p2 ON p2.source_workspace_id = b.workspace_id
                                                           AND p2.merged_at IS NULL
                              WHERE p2.target_workspace_id = t.ws)
                    OR (SELECT count(*) FROM spaces x WHERE x.workspace_id = t.ws)
                       <> COALESCE((SELECT n FROM kb_merge_backup.pre_counts_target
                                     WHERE run_id=run AND workspace_id = t.ws AND table_name = 'spaces'), 0)
                          + (SELECT count(*) FROM kb_merge_backup.spaces b
                               JOIN kb_merge_backup.plan p2 ON p2.source_workspace_id = b.workspace_id
                                                           AND p2.merged_at IS NULL
                              WHERE p2.target_workspace_id = t.ws)) d)),

        -- (5) 移したものが「見える」状態にあること
        --     統合先に既定の役割が 1 つも無ければ、共有スペースでも誰にも見えない
        ('既定の役割が無い統合先', 0::bigint,
            (SELECT count(*) FROM (SELECT DISTINCT target_workspace_id AS ws FROM kb_merge_backup.plan WHERE merged_at IS NULL) t
              WHERE NOT EXISTS (SELECT 1 FROM workspace_grants g WHERE g.workspace_id = t.ws))),
        --     統合した人が統合先で役割を持っていること（主体だけあって役割が無い＝何も見えない、を防ぐ）
        ('役割の無い統合先の主体', 0::bigint,
            (SELECT count(*) FROM kb_merge_backup.principal_map m
              WHERE m.run_id = run
                AND NOT EXISTS (SELECT 1 FROM workspace_grants g
                                 WHERE g.workspace_id = m.target_workspace_id AND g.principal_id = m.target_principal_id)
                AND NOT EXISTS (SELECT 1 FROM space_grants sg
                                 WHERE sg.workspace_id = m.target_workspace_id AND sg.principal_id = m.target_principal_id)))
        ) AS t(label, expected, actual)
    LOOP
        IF v.expected IS DISTINCT FROM v.actual THEN
            problems := problems || format(E'\n  - %s: 期待 %s / 実際 %s', v.label, v.expected, v.actual);
        END IF;
    END LOOP;

    IF problems <> '' THEN
        RAISE EXCEPTION E'0024 の検証に失敗した。トランザクションを巻き戻す:%', problems;
    END IF;
END
$verify$;

-- 統合の要約と、人が判断すべき残り物を報告する。
DO $report$
DECLARE
    run      bigint := currval('kb_merge_backup.run_seq');
    n_ws     bigint;
    n_space  bigint;
    n_page   bigint;
    n_pr     bigint;
    n_sg     bigint;
    privates text;
    stray    text;
BEGIN
    SELECT count(*) INTO n_ws FROM kb_merge_backup.plan WHERE merged_at IS NULL;
    IF n_ws = 0 THEN
        RAISE NOTICE '0024: 統合対象なし（すべて適用済み）。1 行も変更していない';
        RETURN;
    END IF;
    SELECT count(*) INTO n_space FROM kb_merge_backup.spaces b
      JOIN kb_merge_backup.plan p ON p.source_workspace_id=b.workspace_id AND p.merged_at IS NULL;
    SELECT count(*) INTO n_page FROM kb_merge_backup.pages b
      JOIN kb_merge_backup.plan p ON p.source_workspace_id=b.workspace_id AND p.merged_at IS NULL;
    SELECT count(*) INTO n_pr FROM kb_merge_backup.created_principal   WHERE run_id = run;
    SELECT count(*) INTO n_sg FROM kb_merge_backup.created_space_grant WHERE run_id = run;
    RAISE NOTICE '0024: ワークスペース % 個を統合（spaces % / pages %）。統合先に作った主体 % / スペースへ置き直した役割 %',
        n_ws, n_space, n_page, n_pr, n_sg;

    -- private のまま移したスペースは会社の全員には見えない。見え方は移行では変えない。
    -- 入れ物を移すことと公開範囲を広げることを 1 回の操作に混ぜると、片方だけ戻せなくなる。
    SELECT string_agg(format(E'\n  - %s（%s）', s.name, s.id), '') INTO privates
      FROM spaces s
      JOIN kb_merge_backup.spaces b ON b.id = s.id
      JOIN kb_merge_backup.plan p ON p.source_workspace_id = b.workspace_id AND p.merged_at IS NULL
     WHERE s.visibility = 'private';
    IF privates IS NOT NULL THEN
        RAISE NOTICE
            E'統合したスペースのうち visibility=private のものは会社の全員には見えない（移行では見え方を変えない）。開くならこの移行とは別に UPDATE spaces SET visibility=''workspace'' WHERE id IN (...) で行うこと:%',
            privates;
    END IF;

    -- 空にした個人ワークスペースを users.workspace_id が指していないか
    SELECT string_agg(format(E'\n  - user %s → %s', u.id, u.workspace_id), '') INTO stray
      FROM users u JOIN kb_merge_backup.plan p ON p.source_workspace_id = u.workspace_id
     WHERE p.merged_at IS NULL;
    IF stray IS NOT NULL THEN
        RAISE NOTICE
            E'users.workspace_id が統合元を指している（次の起動で tenant_bridge が会社ワークスペースへ直す）:%',
            stray;
    END IF;
END
$report$;

-- ============================================================================
-- 10. 済みの印（これ以降、再実行しても対象 0 件になる）
-- ============================================================================
UPDATE kb_merge_backup.plan SET merged_at = now() WHERE merged_at IS NULL;

COMMIT;

-- ============================================================================
-- 統合が定着し、切り戻す気が無くなったら（数週間後を想定）:
--
--   DROP SCHEMA kb_merge_backup CASCADE;
--
--   空になった個人ワークスペースの行も消すなら、消す前に必ず中身が 0 件であることを確かめる。
--   DELETE は spaces / principals へ CASCADE するので、この確認なしに流さないこと:
--
--   SELECT w.id, w.name FROM workspaces w
--    WHERE w.is_active = false
--      AND NOT EXISTS (SELECT 1 FROM spaces s     WHERE s.workspace_id = w.id)
--      AND NOT EXISTS (SELECT 1 FROM principals p WHERE p.workspace_id = w.id)
--      AND NOT EXISTS (SELECT 1 FROM users u      WHERE u.workspace_id = w.id)
--      AND NOT EXISTS (SELECT 1 FROM companies c  WHERE c.workspace_id = w.id);
-- ============================================================================
