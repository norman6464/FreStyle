-- 個人ワークスペースのノートを、会社ワークスペースへ統合する。
--
-- 背景:
--   会社ごとのワークスペースは起動時のバックフィル（database.BackfillWorkspacesFromCompanies）が
--   作り、companies.workspace_id / users.workspace_id がそれを指す。一方で実際のノートは、
--   利用者が UI から自分で作った「個人ワークスペース」（どの会社にも紐付かない workspaces 行）に
--   溜まっていた。「同じ会社なら全員が見える」修正は会社ワークスペースにしか効かないため、
--   全員が見に行く先は空で、中身はどこにも出てこない状態になっている。
--
--   そこで個人ワークスペースの spaces を、会社ワークスペースの中のスペースとして移す。
--   スペースは visibility='workspace' なら会社全体の grant（editor）が届くので、
--   移した時点で同じ会社の全員に見えるようになる。
--   行を作り直さず workspace_id だけを付け替えるので、spaces / pages / blocks の id は変わらない
--   （共有済みの URL がそのまま生きる）。
--
-- 方針:
--   spaces を付け替え、配下（pages / blocks / page_paths / …）はすべて明示 UPDATE で追従させる。
--   このスキーマの FK に ON UPDATE CASCADE は 1 本も無い（すべて ON UPDATE NO ACTION）ので、
--   親を動かせば子が宙に浮き、子を先に動かせば行き先の親がまだ無い。どちらの順序でも
--   FK 違反になる。DEFERRABLE な制約も 1 つも無いため SET CONSTRAINTS ALL DEFERRED も効かない。
--
--   使えるのは「NO ACTION の検査は文の末尾で走る」という性質。データ変更 CTE（WITH ... UPDATE）は
--   全体で 1 文なので、関係する表をまとめて 1 文で更新すれば、検査が走る時点では
--   親も子も新しい workspace_id に揃っている。
--
--   principals を親とする FK は 6 本ある（page_restrictions / principal_members ×2 /
--   share_links / space_grants / workspace_grants）。**6 本すべての子行を同じ文で
--   始末しないと落ちる。** workspace_grants だけを別の文に残すと、user 以外の主体
--   （group / space_all / share_link）を動かした瞬間にその行が宙に浮く（下の 6 の
--   dropped_workspace_grants を参照）。
--
-- 統合先の決め方（ハードコードしない）:
--   統合先は companies.workspace_id から引く。統合元の user 主体（principals.kind='user'）を
--   users → companies とたどり、その全員が同じ会社に属していればその会社のワークスペースへ移す。
--   会社が割れている / 会社に属さないメンバーが混じっている個人ワークスペースは移さず、
--   そこに中身が残っていれば RAISE EXCEPTION で止める（黙って取り残さない）。
--
-- 冪等:
--   統合済みの状態で流すと、統合元の候補が 0 件になりすべての UPDATE / DELETE が 0 件で終わる。
--   ただしこれは「1 回きりの片付け」であって恒常的な規則ではない。適用後に誰かが新しく
--   ワークスペースを作ってから再実行すると、そのワークスペースも統合対象になる（下の
--   「統合元の対応表」の条件を満たすため）。再実行の前に NOTICE の対象一覧を必ず確認すること。
--
-- トランザクション: 適用は psql -f で行い外部のトランザクション管理は無いため、
-- この中で BEGIN / COMMIT して原子性を確保する（既存 migration と同じ方式）。

BEGIN;

-- ---------------------------------------------------------------------------
-- 1. 統合元 → 統合先の対応表
-- ---------------------------------------------------------------------------
--
-- NOT IN のサブクエリには必ず workspace_id IS NOT NULL を付ける。companies.workspace_id は
-- NULL 可で、NULL が 1 つでも混じると NOT IN 全体が NULL（＝ 常に偽）になり、
-- 統合元が 1 件も選ばれないまま「成功」してしまう。
CREATE TEMP TABLE kb_merge_map (
    source_workspace_id uuid PRIMARY KEY,
    target_workspace_id uuid NOT NULL
) ON COMMIT DROP;

INSERT INTO kb_merge_map (source_workspace_id, target_workspace_id)
-- uuid には min() が無いので、重複を畳んだ配列の先頭を取る（HAVING で 1 件に限っている）。
SELECT p.workspace_id, (array_agg(DISTINCT c.workspace_id))[1]
FROM principals p
JOIN users u ON u.id = p.user_id
JOIN companies c ON c.id = u.company_id AND c.workspace_id IS NOT NULL
WHERE p.kind = 'user'
  AND p.workspace_id NOT IN (
      SELECT workspace_id FROM companies WHERE workspace_id IS NOT NULL
  )
GROUP BY p.workspace_id
-- メンバー全員が同じ 1 つの会社に属していること。
HAVING count(DISTINCT c.workspace_id) = 1
   -- かつ、会社に辿り着けないメンバー（company_id が NULL / 0 / 会社側が未バックフィル）が
   -- 混じっていないこと。その人の主体を統合先で作れず、権限の移し先が決まらないため。
   AND count(*) = (
       SELECT count(*) FROM principals q
       WHERE q.workspace_id = p.workspace_id AND q.kind = 'user'
   );

-- 統合先を決められなかったワークスペースに中身が残っていたら止める。
-- 空のまま残るのは構わない（誰の物か決まらない入れ物を勝手に消さない）が、
-- ノートが取り残されるのは「移行が終わった」と言えない。
DO $$
DECLARE stranded text;
BEGIN
    SELECT string_agg(format('%s（%s）', w.id, w.name), ', ' ORDER BY w.id)
      INTO stranded
    FROM workspaces w
    WHERE w.id NOT IN (SELECT workspace_id FROM companies WHERE workspace_id IS NOT NULL)
      AND w.id NOT IN (SELECT source_workspace_id FROM kb_merge_map)
      AND (EXISTS (SELECT 1 FROM spaces s WHERE s.workspace_id = w.id)
        OR EXISTS (SELECT 1 FROM pages g WHERE g.workspace_id = w.id));
    IF stranded IS NOT NULL THEN
        RAISE EXCEPTION '統合先を決められないワークスペースに中身が残っています: %', stranded;
    END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 2. 移行前の期待値（移行後にこの数と突き合わせる）
-- ---------------------------------------------------------------------------
CREATE TEMP TABLE kb_merge_expected ON COMMIT DROP AS
SELECT t.target_workspace_id,
       (SELECT count(*) FROM spaces s WHERE s.workspace_id = t.target_workspace_id)
     + (SELECT count(*) FROM spaces s JOIN kb_merge_map m ON m.source_workspace_id = s.workspace_id
        WHERE m.target_workspace_id = t.target_workspace_id) AS expected_spaces,
       (SELECT count(*) FROM pages g WHERE g.workspace_id = t.target_workspace_id)
     + (SELECT count(*) FROM pages g JOIN kb_merge_map m ON m.source_workspace_id = g.workspace_id
        WHERE m.target_workspace_id = t.target_workspace_id) AS expected_pages,
       (SELECT count(*) FROM blocks b WHERE b.workspace_id = t.target_workspace_id)
     + (SELECT count(*) FROM blocks b JOIN kb_merge_map m ON m.source_workspace_id = b.workspace_id
        WHERE m.target_workspace_id = t.target_workspace_id) AS expected_blocks
FROM (SELECT DISTINCT target_workspace_id FROM kb_merge_map) t;

-- 全体の総数。1 行も失われていないこと（どこかへ紛れ込んでいないこと）を最後に確かめる。
CREATE TEMP TABLE kb_merge_totals ON COMMIT DROP AS
SELECT (SELECT count(*) FROM spaces)            AS spaces,
       (SELECT count(*) FROM pages)             AS pages,
       (SELECT count(*) FROM blocks)            AS blocks,
       (SELECT count(*) FROM page_paths)        AS page_paths,
       (SELECT count(*) FROM page_snapshots)    AS page_snapshots,
       (SELECT count(*) FROM page_restrictions) AS page_restrictions,
       (SELECT count(*) FROM page_allow_lists)  AS page_allow_lists,
       (SELECT count(*) FROM share_links)       AS share_links;

-- ---------------------------------------------------------------------------
-- 3. 統合先にメンバー（user 主体）を用意する
-- ---------------------------------------------------------------------------
--
-- principals は (workspace_id, user_id) WHERE kind='user' が部分 UNIQUE なので、
-- 統合元の user 主体をそのまま付け替えると、統合先に同じ人の主体が既にある場合に衝突する。
-- そこで「統合先の主体へ寄せる（統合元の主体は捨てる）」に一本化する。
-- 統合先にまだ無い人はここで作る。
--
-- id は UUIDv7 を組み立てる（先頭 48bit = ミリ秒、version=7、variant=10xx）。
WITH src_user AS (
    SELECT DISTINCT m.target_workspace_id AS workspace_id, p.user_id
    FROM principals p
    JOIN kb_merge_map m ON m.source_workspace_id = p.workspace_id
    WHERE p.kind = 'user'
), created AS (
    INSERT INTO principals (id, workspace_id, kind, user_id)
    SELECT (
        lpad(to_hex((extract(epoch FROM clock_timestamp()) * 1000)::bigint), 12, '0')
        || '7' || lpad(to_hex((random() * 4095)::int), 3, '0')
        || to_hex(8 + (random() * 3)::int) || lpad(to_hex((random() * 4095)::int), 3, '0')
        || lpad(to_hex((random() * 281474976710655)::bigint), 12, '0')
    )::uuid, s.workspace_id, 'user', s.user_id
    FROM src_user s
    ON CONFLICT (workspace_id, user_id) WHERE kind = 'user' DO NOTHING
    RETURNING id, workspace_id
)
-- 役割を与えるのは「ここで主体を新しく作った人」だけ（RETURNING は実際に INSERT した行だけ返す）。
-- 既に主体がある人には触らない — 触ると admin が取り消した役割が復活する。
INSERT INTO workspace_grants (workspace_id, principal_id, "role")
SELECT c.workspace_id, c.id, 'editor' FROM created c
ON CONFLICT (workspace_id, principal_id) DO NOTHING;

-- ---------------------------------------------------------------------------
-- 4. 主体の対応表（統合元の principal → 統合後に参照すべき principal）
-- ---------------------------------------------------------------------------
CREATE TEMP TABLE kb_principal_map (
    source_workspace_id uuid NOT NULL,
    target_workspace_id uuid NOT NULL,
    source_principal_id uuid PRIMARY KEY,
    target_principal_id uuid NOT NULL
) ON COMMIT DROP;

INSERT INTO kb_principal_map
SELECT m.source_workspace_id, m.target_workspace_id, p.id,
       CASE WHEN p.kind = 'user' THEN t.id ELSE p.id END
FROM principals p
JOIN kb_merge_map m ON m.source_workspace_id = p.workspace_id
LEFT JOIN principals t
       ON t.workspace_id = m.target_workspace_id
      AND t.kind = 'user'
      AND t.user_id = p.user_id
WHERE p.kind <> 'user' OR t.id IS NOT NULL;

DO $$
DECLARE missing bigint;
BEGIN
    SELECT count(*) INTO missing
    FROM principals p
    JOIN kb_merge_map m ON m.source_workspace_id = p.workspace_id
    WHERE NOT EXISTS (
        SELECT 1 FROM kb_principal_map k WHERE k.source_principal_id = p.id
    );
    IF missing > 0 THEN
        RAISE EXCEPTION '統合先の主体を用意できなかった principal が % 件あります', missing;
    END IF;
END
$$;

-- グループ名はワークスペース内で一意（uq_principals_group_name）。統合先に同名グループが
-- あると付け替えで衝突する。統合の意味は人が決めることなので、ここでは止めて知らせる。
DO $$
DECLARE dup text;
BEGIN
    SELECT string_agg(DISTINCT src.name, ', ') INTO dup
    FROM principals src
    JOIN kb_merge_map m ON m.source_workspace_id = src.workspace_id
    JOIN principals tgt
      ON tgt.workspace_id = m.target_workspace_id
     AND tgt.kind = 'group' AND tgt.name = src.name
    WHERE src.kind = 'group';
    IF dup IS NOT NULL THEN
        RAISE EXCEPTION 'グループ名が統合先と衝突します（統合方法を決めてから再実行）: %', dup;
    END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 5. スペースの key 衝突を先に解く
-- ---------------------------------------------------------------------------
--
-- spaces は UNIQUE (workspace_id, "key")。衝突する分だけ、スペース ID から必ず一意になる
-- key を作り直す。長さは left(old,30) + '-' + uuid の 32 桁 = 63 で ck_spaces_key_len に収まる。
CREATE TEMP TABLE kb_space_move (
    space_id            uuid PRIMARY KEY,
    source_workspace_id uuid NOT NULL,
    target_workspace_id uuid NOT NULL,
    old_key             varchar(64) NOT NULL,
    new_key             varchar(64) NOT NULL
) ON COMMIT DROP;

INSERT INTO kb_space_move
SELECT s.id, s.workspace_id, m.target_workspace_id, s."key", s."key"
FROM spaces s
JOIN kb_merge_map m ON m.source_workspace_id = s.workspace_id;

UPDATE kb_space_move mv
SET new_key = left(mv.old_key, 30) || '-' || replace(mv.space_id::text, '-', '')
WHERE EXISTS (
        SELECT 1 FROM spaces t
        WHERE t.workspace_id = mv.target_workspace_id
          AND t."key" = mv.old_key AND t.id <> mv.space_id
      )
   OR EXISTS (
        SELECT 1 FROM kb_space_move o
        WHERE o.target_workspace_id = mv.target_workspace_id
          AND o.space_id <> mv.space_id AND o.old_key = mv.old_key
      );

DO $$
DECLARE conflicted bigint;
BEGIN
    SELECT count(*) INTO conflicted
    FROM kb_space_move mv
    WHERE EXISTS (
            SELECT 1 FROM spaces t
            WHERE t.workspace_id = mv.target_workspace_id
              AND t."key" = mv.new_key AND t.id <> mv.space_id
          )
       OR EXISTS (
            SELECT 1 FROM kb_space_move o
            WHERE o.target_workspace_id = mv.target_workspace_id
              AND o.space_id <> mv.space_id AND o.new_key = mv.new_key
          );
    IF conflicted > 0 THEN
        RAISE EXCEPTION 'スペースの key 衝突を解消できませんでした（% 件）', conflicted;
    END IF;
END
$$;

-- ---------------------------------------------------------------------------
-- 5-bis. 統合元のワークスペース全体の役割を控える（本体の前に必ず取る）
-- ---------------------------------------------------------------------------
--
-- workspace_grants は principals への複合 FK（fk_workspace_grants_principal）を持ち、
-- ON UPDATE は NO ACTION。本体で user 以外の主体（group / space_all / share_link）を
-- 統合先へ動かすと、その主体を指したままの workspace_grants が宙に浮き、文の末尾の検査で
--   update or delete on table "principals" violates foreign key constraint
--   "fk_workspace_grants_principal" on table "workspace_grants"
-- になって移行全体が落ちる。GrantWorkspaceRoleUseCase は主体の種類を問わない
-- （実在とテナント一致しか見ない）ので、グループにワークスペース全体の役割を与えた状態は
-- UI から普通に作れる。
--
-- そこで「本体と同じ文の中で統合元の workspace_grants を消す」ことで、検査の時点で
-- 宙に浮く子行を無くす。消す前の内容はここで控え、あとの役割の翻訳はこの控えから行う。
CREATE TEMP TABLE kb_source_workspace_grants ON COMMIT DROP AS
SELECT wg.workspace_id, wg.principal_id, wg."role"
FROM workspace_grants wg
JOIN kb_merge_map m ON m.source_workspace_id = wg.workspace_id;

-- ---------------------------------------------------------------------------
-- 6. 付け替え（ここだけは 1 文でまとめて行う）
-- ---------------------------------------------------------------------------
--
-- FK の検査は文の末尾でまとめて走る。表ごとに文を分けると、その時点で親か子のどちらかが
-- 必ず宙に浮いて FK 違反になる。
--
-- updated_at は動かさない（テナントの付け替えは人の編集ではないため。pages.updated_at は
-- 画面の「最近更新」の並びに使われる）。spaces.visibility も動かさない
-- （private を workspace へ倒すと、本人だけに見えていたものが会社全員に開く）。
CREATE TEMP TABLE kb_move_counts (
    spaces bigint, pages bigint, blocks bigint, page_paths bigint,
    page_restrictions bigint, page_allow_lists bigint, share_links bigint,
    principals bigint, principal_members bigint, space_grants bigint,
    workspace_grants bigint
) ON COMMIT DROP;

WITH moved_spaces AS (
    UPDATE spaces s
    SET workspace_id = mv.target_workspace_id,
        "key"        = mv.new_key
    FROM kb_space_move mv
    WHERE s.id = mv.space_id
    RETURNING s.id
), moved_pages AS (
    UPDATE pages g
    SET workspace_id = m.target_workspace_id
    FROM kb_merge_map m
    WHERE g.workspace_id = m.source_workspace_id
    RETURNING g.id
), moved_blocks AS (
    UPDATE blocks b
    SET workspace_id = m.target_workspace_id
    FROM kb_merge_map m
    WHERE b.workspace_id = m.source_workspace_id
    RETURNING b.id
), moved_page_paths AS (
    UPDATE page_paths pp
    SET workspace_id = m.target_workspace_id
    FROM kb_merge_map m
    WHERE pp.workspace_id = m.source_workspace_id
    RETURNING pp.page_id
), moved_restrictions AS (
    UPDATE page_restrictions pr
    SET workspace_id = k.target_workspace_id,
        principal_id = k.target_principal_id
    FROM kb_principal_map k
    WHERE pr.workspace_id = k.source_workspace_id
      AND pr.principal_id = k.source_principal_id
    RETURNING pr.page_id
), moved_allow_lists AS (
    UPDATE page_allow_lists pal
    SET workspace_id = m.target_workspace_id
    FROM kb_merge_map m
    WHERE pal.workspace_id = m.source_workspace_id
    RETURNING pal.page_id
), moved_share_links AS (
    UPDATE share_links sl
    SET workspace_id = m.target_workspace_id
    FROM kb_merge_map m
    WHERE sl.workspace_id = m.source_workspace_id
    RETURNING sl.id
), moved_principals AS (
    -- user 主体は動かさない（統合先の主体へ寄せたので、この後で消す）。
    UPDATE principals p
    SET workspace_id = m.target_workspace_id
    FROM kb_merge_map m
    WHERE p.workspace_id = m.source_workspace_id
      AND p.kind <> 'user'
    RETURNING p.id
), moved_principal_members AS (
    UPDATE principal_members pm
    SET workspace_id         = kg.target_workspace_id,
        group_principal_id   = kg.target_principal_id,
        member_principal_id  = km.target_principal_id
    FROM kb_principal_map kg, kb_principal_map km
    WHERE pm.workspace_id        = kg.source_workspace_id
      AND pm.group_principal_id  = kg.source_principal_id
      AND pm.workspace_id        = km.source_workspace_id
      AND pm.member_principal_id = km.source_principal_id
    RETURNING pm.group_principal_id
), moved_space_grants AS (
    UPDATE space_grants sg
    SET workspace_id = k.target_workspace_id,
        principal_id = k.target_principal_id
    FROM kb_principal_map k
    WHERE sg.workspace_id = k.source_workspace_id
      AND sg.principal_id = k.source_principal_id
    RETURNING sg.space_id
), dropped_workspace_grants AS (
    -- 統合元のワークスペース全体の役割は、そのまま持っていくと会社全体の admin が
    -- 人数分生まれるので捨てる（下の 7 でスペース単位の役割へ翻訳する。控えは 5-bis）。
    -- **この DELETE は本体と同じ文でなければならない。** 別の文にすると、
    -- 本体が user 以外の主体を動かした時点でこの表の行が宙に浮き、
    -- fk_workspace_grants_principal の検査（NO ACTION・文の末尾）に当たって落ちる。
    DELETE FROM workspace_grants wg
    USING kb_merge_map m
    WHERE wg.workspace_id = m.source_workspace_id
    RETURNING wg.principal_id
)
INSERT INTO kb_move_counts
SELECT (SELECT count(*) FROM moved_spaces),
       (SELECT count(*) FROM moved_pages),
       (SELECT count(*) FROM moved_blocks),
       (SELECT count(*) FROM moved_page_paths),
       (SELECT count(*) FROM moved_restrictions),
       (SELECT count(*) FROM moved_allow_lists),
       (SELECT count(*) FROM moved_share_links),
       (SELECT count(*) FROM moved_principals),
       (SELECT count(*) FROM moved_principal_members),
       (SELECT count(*) FROM moved_space_grants),
       (SELECT count(*) FROM dropped_workspace_grants);

-- ---------------------------------------------------------------------------
-- 7. ワークスペース全体の役割を、移したスペースの役割へ翻訳する
-- ---------------------------------------------------------------------------
--
-- 個人ワークスペースの作成者は自分のワークスペースの admin なので、そのまま持っていくと
-- 全員が会社ワークスペース全体の admin になる。一方で「自分が作ったスペースは自分で
-- 管理できる」は残したいので、移したスペースそれぞれに同じ主体・同じ役割の
-- space_grant を張る（控え kb_source_workspace_grants から読む。本体で元表は空にした）。
INSERT INTO space_grants (workspace_id, space_id, principal_id, "role")
SELECT mv.target_workspace_id, mv.space_id, k.target_principal_id, wg."role"
FROM kb_source_workspace_grants wg
JOIN kb_principal_map k
  ON k.source_workspace_id = wg.workspace_id
 AND k.source_principal_id = wg.principal_id
JOIN kb_space_move mv ON mv.source_workspace_id = wg.workspace_id
ON CONFLICT (workspace_id, space_id, principal_id) DO UPDATE
SET "role" = EXCLUDED."role",
    updated_at = now()
-- 既に強い（か同じ）役割が張られているなら書き換えない（順位は domain.GrantRole.Rank と同じ）。
WHERE array_position(ARRAY['viewer', 'commenter', 'editor', 'admin'], EXCLUDED."role")
    > array_position(ARRAY['viewer', 'commenter', 'editor', 'admin'], space_grants."role");

-- ---------------------------------------------------------------------------
-- 8. 統合元が空になったことを確かめてから、主体とワークスペースを消す
-- ---------------------------------------------------------------------------
DO $$
DECLARE leftovers text;
BEGIN
    SELECT string_agg(t.label || '=' || t.n, ', ' ORDER BY t.label) INTO leftovers
    FROM (
        SELECT 'spaces' AS label, count(*) AS n FROM spaces s
          JOIN kb_merge_map m ON m.source_workspace_id = s.workspace_id
        UNION ALL SELECT 'pages', count(*) FROM pages g
          JOIN kb_merge_map m ON m.source_workspace_id = g.workspace_id
        UNION ALL SELECT 'blocks', count(*) FROM blocks b
          JOIN kb_merge_map m ON m.source_workspace_id = b.workspace_id
        UNION ALL SELECT 'page_paths', count(*) FROM page_paths pp
          JOIN kb_merge_map m ON m.source_workspace_id = pp.workspace_id
        UNION ALL SELECT 'page_restrictions', count(*) FROM page_restrictions pr
          JOIN kb_merge_map m ON m.source_workspace_id = pr.workspace_id
        UNION ALL SELECT 'page_allow_lists', count(*) FROM page_allow_lists pal
          JOIN kb_merge_map m ON m.source_workspace_id = pal.workspace_id
        UNION ALL SELECT 'share_links', count(*) FROM share_links sl
          JOIN kb_merge_map m ON m.source_workspace_id = sl.workspace_id
        UNION ALL SELECT 'workspace_grants', count(*) FROM workspace_grants wg
          JOIN kb_merge_map m ON m.source_workspace_id = wg.workspace_id
        UNION ALL SELECT 'space_grants', count(*) FROM space_grants sg
          JOIN kb_merge_map m ON m.source_workspace_id = sg.workspace_id
        UNION ALL SELECT 'principal_members', count(*) FROM principal_members pm
          JOIN kb_merge_map m ON m.source_workspace_id = pm.workspace_id
        UNION ALL SELECT 'principals(kind<>user)', count(*) FROM principals p
          JOIN kb_merge_map m ON m.source_workspace_id = p.workspace_id
          WHERE p.kind <> 'user'
    ) t
    WHERE t.n > 0;
    IF leftovers IS NOT NULL THEN
        RAISE EXCEPTION '統合元に行が残っています（移行が不完全）: %', leftovers;
    END IF;
END
$$;

-- 残っているのは統合先へ寄せた分の user 主体だけ。参照はすべて移し終えているので、
-- ここでの ON DELETE CASCADE は 1 行も巻き込まない（直前の検査がそれを保証している）。
DELETE FROM principals p
USING kb_merge_map m
WHERE p.workspace_id = m.source_workspace_id;

-- users.workspace_id / companies.workspace_id から指されていたら消せない
-- （fk_users_workspace / fk_companies_workspace は NO ACTION）。何が指しているかを知らせる。
DO $$
DECLARE referenced text;
BEGIN
    SELECT string_agg(DISTINCT u.id::text, ', ') INTO referenced
    FROM users u JOIN kb_merge_map m ON m.source_workspace_id = u.workspace_id;
    IF referenced IS NOT NULL THEN
        RAISE EXCEPTION
            '統合元ワークスペースを users.workspace_id が参照しています（user id: %）。'
            '所属の付け替えを先に決めてください', referenced;
    END IF;
END
$$;

DELETE FROM workspaces w
USING kb_merge_map m
WHERE w.id = m.source_workspace_id;

-- ---------------------------------------------------------------------------
-- 9. 検証
-- ---------------------------------------------------------------------------
DO $$
DECLARE
    rec  record;
    tot  record;
    n    bigint;
BEGIN
    FOR rec IN SELECT * FROM kb_merge_expected LOOP
        SELECT count(*) INTO n FROM spaces WHERE workspace_id = rec.target_workspace_id;
        IF n <> rec.expected_spaces THEN
            RAISE EXCEPTION '統合先 % の spaces が % 件（期待 % 件）',
                rec.target_workspace_id, n, rec.expected_spaces;
        END IF;
        SELECT count(*) INTO n FROM pages WHERE workspace_id = rec.target_workspace_id;
        IF n <> rec.expected_pages THEN
            RAISE EXCEPTION '統合先 % の pages が % 件（期待 % 件）',
                rec.target_workspace_id, n, rec.expected_pages;
        END IF;
        SELECT count(*) INTO n FROM blocks WHERE workspace_id = rec.target_workspace_id;
        IF n <> rec.expected_blocks THEN
            RAISE EXCEPTION '統合先 % の blocks が % 件（期待 % 件）',
                rec.target_workspace_id, n, rec.expected_blocks;
        END IF;
    END LOOP;

    SELECT * INTO tot FROM kb_merge_totals;
    IF (SELECT count(*) FROM spaces) <> tot.spaces
       OR (SELECT count(*) FROM pages) <> tot.pages
       OR (SELECT count(*) FROM blocks) <> tot.blocks
       OR (SELECT count(*) FROM page_paths) <> tot.page_paths
       OR (SELECT count(*) FROM page_snapshots) <> tot.page_snapshots
       OR (SELECT count(*) FROM page_restrictions) <> tot.page_restrictions
       OR (SELECT count(*) FROM page_allow_lists) <> tot.page_allow_lists
       OR (SELECT count(*) FROM share_links) <> tot.share_links
    THEN
        RAISE EXCEPTION
            '移行の前後で総数が変わりました（spaces %→%, pages %→%, blocks %→%, page_paths %→%, snapshots %→%, restrictions %→%, allow_lists %→%, share_links %→%）',
            tot.spaces, (SELECT count(*) FROM spaces),
            tot.pages, (SELECT count(*) FROM pages),
            tot.blocks, (SELECT count(*) FROM blocks),
            tot.page_paths, (SELECT count(*) FROM page_paths),
            tot.page_snapshots, (SELECT count(*) FROM page_snapshots),
            tot.page_restrictions, (SELECT count(*) FROM page_restrictions),
            tot.page_allow_lists, (SELECT count(*) FROM page_allow_lists),
            tot.share_links, (SELECT count(*) FROM share_links);
    END IF;

    SELECT count(*) INTO n FROM workspaces w JOIN kb_merge_map m ON m.source_workspace_id = w.id;
    IF n > 0 THEN
        RAISE EXCEPTION '統合元のワークスペースが % 件残っています', n;
    END IF;

    SELECT count(*) INTO n
    FROM pages g LEFT JOIN spaces s ON s.id = g.space_id AND s.workspace_id = g.workspace_id
    WHERE s.id IS NULL;
    IF n > 0 THEN
        RAISE EXCEPTION 'スペースとワークスペースが食い違うページが % 件あります', n;
    END IF;
END
$$;

-- 何をしたかを残す（適用ログに出る）。
DO $$
DECLARE c record; sources bigint; privates bigint;
BEGIN
    SELECT count(*) INTO sources FROM kb_merge_map;
    SELECT * INTO c FROM kb_move_counts;
    SELECT count(*) INTO privates
    FROM spaces s JOIN kb_space_move mv ON mv.space_id = s.id
    WHERE s.visibility = 'private';
    RAISE NOTICE
        '統合したワークスペース: % 件 / spaces % / pages % / blocks % / page_paths % / restrictions % / allow_lists % / share_links % / principals(移動) % / principal_members % / space_grants(移動) % / workspace_grants(翻訳して削除) %',
        sources, c.spaces, c.pages, c.blocks, c.page_paths, c.page_restrictions,
        c.page_allow_lists, c.share_links, c.principals, c.principal_members, c.space_grants,
        c.workspace_grants;
    IF privates > 0 THEN
        RAISE NOTICE
            'うち visibility=private のスペースが % 件あります。private のままなので会社の全員には見えません（付与された人だけに見えます）',
            privates;
    END IF;
END
$$;

COMMIT;