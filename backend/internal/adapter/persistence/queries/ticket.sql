-- チケット（段 1: 骨格）のクエリ。knowledge_base.sql と同じ作法。
--
-- 作法（このファイル全体の前提）:
--   - すべての SELECT / UPDATE / DELETE の WHERE に workspace_id を含める。
--   - UPDATE 文には必ず updated_at = now() を明示する。
--   - position（COLLATE "C"）の ORDER BY はバイト順（fracindex と一致）。
--   - tickets への INSERT は CreateTicket 1 本だけ（採番 CTE を含む文でなければ番号が
--     ticket_counters と無関係に振られてしまう。設計 Ⅳ-B のレビュー項目）。

-- =============================================================================
-- ticket_statuses（管理画面）
-- =============================================================================

-- name: HasActiveInitialTicketStatus :one
-- 「有効化済み」の正本判定: 初期状態を持つ現役の状態が 1 つでもあるか。
SELECT EXISTS (
  SELECT 1 FROM ticket_statuses
  WHERE workspace_id = sqlc.arg(workspace_id) AND space_id = sqlc.arg(space_id)
    AND is_initial AND archived_at IS NULL
) AS exists;

-- name: InsertTicketStatus :one
INSERT INTO ticket_statuses
  (id, workspace_id, space_id, name, category, color, "position", is_initial, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now(), now())
RETURNING *;

-- name: GetTicketStatus :one
SELECT * FROM ticket_statuses
WHERE workspace_id = $1 AND space_id = $2 AND id = $3;

-- name: ListTicketStatuses :many
-- sqlc.narg(archived) は bool。呼び出し側は「現役だけ」か「アーカイブ済みだけ」かを
-- 明示的に渡す（NULL で「両方」は扱わない — 管理画面のタブ切り替えに 1 対 1 対応させる）。
SELECT * FROM ticket_statuses
WHERE workspace_id = sqlc.arg(workspace_id) AND space_id = sqlc.arg(space_id)
  AND (archived_at IS NOT NULL) = sqlc.arg(archived)::boolean
ORDER BY "position";

-- name: GetInitialTicketStatus :one
SELECT * FROM ticket_statuses
WHERE workspace_id = $1 AND space_id = $2 AND is_initial AND archived_at IS NULL;

-- name: UpdateTicketStatus :one
UPDATE ticket_statuses
SET name = $4, category = $5, color = $6, updated_at = now()
WHERE workspace_id = $1 AND space_id = $2 AND id = $3
RETURNING *;

-- name: ClearTicketStatusInitial :execrows
-- SetInitialTicketStatus は「旧初期状態を先に false へ倒す → 新しい状態を true にする」の
-- 2 文で、usecase が同一トランザクションで呼ぶ（部分 UNIQUE のため同時に 2 つは作れない。
-- 旧初期状態が無いスペースでは 0 行更新で構わない）。
UPDATE ticket_statuses
SET is_initial = false, updated_at = now()
WHERE workspace_id = $1 AND space_id = $2 AND is_initial AND archived_at IS NULL;

-- name: SetTicketStatusInitial :execrows
UPDATE ticket_statuses
SET is_initial = true, updated_at = now()
WHERE workspace_id = $1 AND space_id = $2 AND id = $3 AND archived_at IS NULL;

-- name: ArchiveTicketStatus :execrows
UPDATE ticket_statuses
SET archived_at = now(), updated_at = now()
WHERE workspace_id = $1 AND space_id = $2 AND id = $3 AND archived_at IS NULL;

-- name: RestoreTicketStatus :execrows
-- position を末尾へ付け直す（呼び出し側が LastActiveTicketStatusPosition から
-- fracindex.Between で採番した値を渡す）。
UPDATE ticket_statuses
SET archived_at = NULL, "position" = $4, updated_at = now()
WHERE workspace_id = $1 AND space_id = $2 AND id = $3 AND archived_at IS NOT NULL;

-- name: CountActiveTicketsByStatus :one
SELECT count(*) FROM tickets
WHERE workspace_id = $1 AND space_id = $2 AND status_id = $3 AND archived_at IS NULL;

-- name: LastActiveTicketStatusPosition :one
-- 現役の状態のうち最後（position 最大）のもの。復元・新規作成の末尾採番に使う。
-- 1 件も無ければ空文字（sqlc は :one で 0 行だと sql.ErrNoRows を返すため、
-- COALESCE で空文字に畳んで「0 行エラー」を避ける）。
SELECT COALESCE(max("position"), '')::text AS "position" FROM ticket_statuses
WHERE workspace_id = $1 AND space_id = $2 AND archived_at IS NULL;

-- =============================================================================
-- ticket_types（管理画面）
-- =============================================================================

-- name: InsertTicketType :one
INSERT INTO ticket_types
  (id, workspace_id, space_id, name, color, hierarchy_level, "position", is_default,
   template_title, template_doc, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now(), now())
RETURNING *;

-- name: GetTicketType :one
SELECT * FROM ticket_types
WHERE workspace_id = $1 AND space_id = $2 AND id = $3;

-- name: ListTicketTypes :many
SELECT * FROM ticket_types
WHERE workspace_id = sqlc.arg(workspace_id) AND space_id = sqlc.arg(space_id)
  AND (archived_at IS NOT NULL) = sqlc.arg(archived)::boolean
ORDER BY "position";

-- name: GetDefaultTicketType :one
SELECT * FROM ticket_types
WHERE workspace_id = $1 AND space_id = $2 AND is_default AND archived_at IS NULL;

-- name: UpdateTicketType :one
UPDATE ticket_types
SET name = $4, color = $5, hierarchy_level = $6,
    template_title = $7, template_doc = $8, updated_at = now()
WHERE workspace_id = $1 AND space_id = $2 AND id = $3
RETURNING *;

-- name: ClearTicketTypeDefault :execrows
UPDATE ticket_types
SET is_default = false, updated_at = now()
WHERE workspace_id = $1 AND space_id = $2 AND is_default AND archived_at IS NULL;

-- name: SetTicketTypeDefault :execrows
UPDATE ticket_types
SET is_default = true, updated_at = now()
WHERE workspace_id = $1 AND space_id = $2 AND id = $3 AND archived_at IS NULL;

-- name: ArchiveTicketType :execrows
UPDATE ticket_types
SET archived_at = now(), updated_at = now()
WHERE workspace_id = $1 AND space_id = $2 AND id = $3 AND archived_at IS NULL;

-- name: RestoreTicketType :execrows
UPDATE ticket_types
SET archived_at = NULL, "position" = $4, updated_at = now()
WHERE workspace_id = $1 AND space_id = $2 AND id = $3 AND archived_at IS NOT NULL;

-- name: CountActiveTicketsByType :one
SELECT count(*) FROM tickets
WHERE workspace_id = $1 AND space_id = $2 AND type_id = $3 AND archived_at IS NULL;

-- name: LastActiveTicketTypePosition :one
SELECT COALESCE(max("position"), '')::text AS "position" FROM ticket_types
WHERE workspace_id = $1 AND space_id = $2 AND archived_at IS NULL;

-- =============================================================================
-- tickets 本体
-- =============================================================================

-- name: CreateTicket :one
-- 採番と INSERT を 1 文の CTE にまとめる（設計 Ⅳ-B）。本番の transaction pooler 越しでも
-- 接続が同じであることが保証され、行ロックで直列化される（20 並行で番号が連続することを
-- 実機で確認済み）。VALUES に 1 を渡すのは初回の初期化（ON CONFLICT で 2 回目以降は +1）。
WITH n AS (
  INSERT INTO ticket_counters (workspace_id, space_id, last_number, updated_at)
  VALUES (sqlc.arg(workspace_id), sqlc.arg(space_id), 1, now())
  ON CONFLICT (workspace_id, space_id)
  DO UPDATE SET last_number = ticket_counters.last_number + 1, updated_at = now()
  RETURNING last_number
)
INSERT INTO tickets
  (id, workspace_id, space_id, number, type_id, status_id, parent_id, title, doc, plain_text,
   priority, start_date, due_date, "position", created_by_user_id, created_at, updated_at)
SELECT
  sqlc.arg(id), sqlc.arg(workspace_id), sqlc.arg(space_id), n.last_number,
  sqlc.arg(type_id), sqlc.arg(status_id), sqlc.narg(parent_id), sqlc.arg(title),
  sqlc.arg(doc), sqlc.arg(plain_text), sqlc.arg(priority),
  sqlc.narg(start_date)::date, sqlc.narg(due_date)::date, sqlc.arg(position),
  sqlc.arg(created_by_user_id), now(), now()
FROM n
RETURNING *;

-- name: GetTicket :one
-- 担当（ticket_assignments）を LEFT JOIN で添える。画面は詳細でも一覧でも担当を出すので、
-- チケット 1 件につき問い合わせを 2 回に分けない（設計 Ⅶ の「詳細（… 担当 …）」）。
-- 担当は 1 人（ticket_id が PK）なので、この JOIN で行が増えることはない。
--
-- ticket_ranks（段 2）を LEFT JOIN して並び順を rank_position として添える。並び順の正本は
-- tickets.position から ticket_ranks.position へ移った（設計 Ⅳ-F）。tickets.position 列は
-- まだ残っているが（段 2 では DROP しない）。LEFT JOIN + COALESCE にしてあるのは、
-- CreateTicket と InsertTicketRank が別の 2 文（CreateTicketUseCase 参照）で、どちらかを
-- 単独で呼ぶ経路（結合テストの直接呼び出し等）があってもチケットが一覧から消えないようにする
-- ため。INNER JOIN だと ticket_ranks 側の行が無いだけでチケットが「無い」と誤認される。
--
-- deleted_at IS NOT NULL のチケットは「無い」と同じ扱いにする（設計 Ⅳ-J: 消えたことにする。
-- archived_at と違い戻す口を持たない）。削除済みチケットを個別に引く経路は
-- FindDeletedTicket に分けてある（RestoreDeletedTicketUseCase 専用）。
SELECT t.*, a.assignee_principal_id, COALESCE(r.position, t."position") AS rank_position FROM tickets t
LEFT JOIN ticket_ranks r ON r.workspace_id = t.workspace_id AND r.ticket_id = t.id AND r.context_kind = 'backlog'
LEFT JOIN ticket_assignments a ON a.workspace_id = t.workspace_id AND a.ticket_id = t.id
WHERE t.workspace_id = $1 AND t.id = $2 AND t.deleted_at IS NULL;

-- name: FindDeletedTicket :one
-- RestoreDeletedTicketUseCase 専用。GetTicket と逆に、削除済み（deleted_at IS NOT NULL）の
-- 行だけを引く（現役の行は見えない — Archive/Restore の archived_at と対称の作法）。
SELECT * FROM tickets
WHERE workspace_id = $1 AND id = $2 AND deleted_at IS NOT NULL;

-- name: GetTicketForUpdate :one
-- 状態変更・親子変更・順位変更の直前にロックする。
SELECT * FROM tickets
WHERE workspace_id = $1 AND id = $2
FOR UPDATE;

-- name: ResolveTicketIDByKey :one
-- spaceKey（小文字。domain.ParseTicketKey が返す）+ number からチケットを引く。
SELECT t.id, t.workspace_id FROM tickets t
JOIN spaces s ON s.workspace_id = t.workspace_id AND s.id = t.space_id
WHERE t.workspace_id = sqlc.arg(workspace_id)
  AND lower(s."key") = sqlc.arg(space_key)
  AND t.number = sqlc.arg(number);

-- name: ListTickets :many
-- status_id / type_id / assignee_principal_id はいずれも sqlc.narg。NULL なら絞らない。
-- ticket_ranks を LEFT JOIN + COALESCE で並び順を rank_position として返す（GetTicket と同じ理由）。
-- deleted_at IS NULL は常に付ける（include_archived の有無に関わらず、削除済みは一覧に出さない）。
SELECT t.*, a.assignee_principal_id, COALESCE(r.position, t."position") AS rank_position FROM tickets t
LEFT JOIN ticket_ranks r ON r.workspace_id = t.workspace_id AND r.ticket_id = t.id AND r.context_kind = 'backlog'
LEFT JOIN ticket_assignments a ON a.workspace_id = t.workspace_id AND a.ticket_id = t.id
WHERE t.workspace_id = sqlc.arg(workspace_id) AND t.space_id = sqlc.arg(space_id)
  AND t.deleted_at IS NULL
  AND (t.archived_at IS NOT NULL) = sqlc.arg(include_archived)::boolean
  AND (sqlc.narg(status_id)::uuid IS NULL OR t.status_id = sqlc.narg(status_id)::uuid)
  AND (sqlc.narg(type_id)::uuid IS NULL OR t.type_id = sqlc.narg(type_id)::uuid)
  AND (
    sqlc.narg(assignee_principal_id)::uuid IS NULL
    OR a.assignee_principal_id = sqlc.narg(assignee_principal_id)::uuid
  )
ORDER BY COALESCE(r.position, t."position");

-- name: ListTicketChildren :many
-- ticket_ranks を LEFT JOIN + COALESCE で並び順を rank_position として返す（同上）。
SELECT t.*, COALESCE(r.position, t."position") AS rank_position FROM tickets t
LEFT JOIN ticket_ranks r ON r.workspace_id = t.workspace_id AND r.ticket_id = t.id AND r.context_kind = 'backlog'
WHERE t.workspace_id = $1 AND t.space_id = $2 AND t.parent_id = $3
  AND t.archived_at IS NULL AND t.deleted_at IS NULL
ORDER BY COALESCE(r.position, t."position");

-- name: UpdateTicket :one
UPDATE tickets
SET type_id = sqlc.arg(type_id), parent_id = sqlc.narg(parent_id), title = sqlc.arg(title),
    doc = sqlc.arg(doc), plain_text = sqlc.arg(plain_text), priority = sqlc.arg(priority),
    start_date = sqlc.narg(start_date)::date, due_date = sqlc.narg(due_date)::date,
    updated_at = now()
WHERE workspace_id = sqlc.arg(workspace_id) AND id = sqlc.arg(id)
RETURNING *;

-- name: ChangeTicketStatus :one
-- closed_at / resolution は usecase が domain.ResolveTicketClosedFields で導出した値を
-- そのまま渡す（ここでは category との整合を判断しない）。
UPDATE tickets
SET status_id = $3, closed_at = $4, resolution = $5, updated_at = now()
WHERE workspace_id = $1 AND id = $2
RETURNING *;

-- name: ArchiveTicket :execrows
UPDATE tickets
SET archived_at = now(), updated_at = now()
WHERE workspace_id = $1 AND id = $2 AND archived_at IS NULL AND deleted_at IS NULL;

-- name: RestoreTicket :execrows
UPDATE tickets
SET archived_at = NULL, "position" = $3, updated_at = now()
WHERE workspace_id = $1 AND id = $2 AND archived_at IS NOT NULL AND deleted_at IS NULL;

-- name: DeleteTicket :execrows
-- 「消えたことにする」（設計 Ⅳ-J）。archived_at と独立の列で、戻す口は
-- RestoreDeletedTicket だけ（一覧・検索・URL 直打ちのどこにも出てこなくなる）。
-- 既に削除済みなら 0 行（呼び出し側は ErrTicketNotFound に畳む。冪等な 404）。
UPDATE tickets
SET deleted_at = now(), updated_at = now()
WHERE workspace_id = $1 AND id = $2 AND deleted_at IS NULL;

-- name: RestoreDeletedTicket :execrows
-- position は末尾へ付け直す（RestoreTicket と同じ理由。削除されていた間に他のチケットの
-- 並びが進んでいる可能性があるため、元の位置は復元しない）。
UPDATE tickets
SET deleted_at = NULL, "position" = $3, updated_at = now()
WHERE workspace_id = $1 AND id = $2 AND deleted_at IS NOT NULL;

-- name: DeleteTicketPageLinksBySourceCascade :exec
-- チケット削除時、その本文からの参照（派生索引）も一緒に「消えたことにする」。
-- 物理削除しないのは、DeleteTicketPageLinksBySource（本文保存時の張り替え）と役割が違うため
-- — こちらは「参照元が消えたので隠す」、あちらは「本文が変わったので作り直す」。
UPDATE ticket_page_links
SET deleted_at = now()
WHERE workspace_id = $1 AND source_ticket_id = $2 AND deleted_at IS NULL;

-- name: DeleteTicketTicketLinksBySourceCascade :exec
UPDATE ticket_ticket_links
SET deleted_at = now()
WHERE workspace_id = $1 AND source_ticket_id = $2 AND deleted_at IS NULL;

-- name: CountActiveTicketChildren :one
SELECT count(*) FROM tickets
WHERE workspace_id = $1 AND parent_id = $2 AND archived_at IS NULL;

-- name: ListTicketParentChain :many
-- 親を根まで辿る（自分は含まない、根に近い順）。最大 3 段の設計なので再帰は浅く終わるが、
-- 誤ったデータで循環していても RECURSIVE は無限ループしない（訪問済み id を UNION の
-- 重複排除では止められないため、深さで打ち切る）。
WITH RECURSIVE chain AS (
  SELECT t.*, 0 AS depth
  FROM tickets t
  WHERE t.workspace_id = sqlc.arg(workspace_id) AND t.id = sqlc.arg(ticket_id)
  UNION ALL
  SELECT p.*, c.depth + 1
  FROM tickets p
  JOIN chain c ON p.workspace_id = c.workspace_id AND p.id = c.parent_id
  WHERE c.depth < 10
)
SELECT id, workspace_id, space_id, number, type_id, status_id, parent_id, title, doc,
  plain_text, priority, start_date, due_date, "position", closed_at, resolution,
  created_by_user_id, archived_at, deleted_at, created_at, updated_at
FROM chain
WHERE depth > 0
ORDER BY depth DESC;

-- name: LastActiveTicketPosition :one
-- tickets.position 自体は段 2 で並び順の正本ではなくなったが、CreateTicket の INSERT が
-- NOT NULL 列を埋めるためにまだこれを呼ぶ（列は残す。読み手は誰も居ない・書き手だけ残る）。
SELECT COALESCE(max("position"), '')::text AS "position" FROM tickets
WHERE workspace_id = $1 AND space_id = $2 AND archived_at IS NULL AND deleted_at IS NULL;

-- name: FindActiveTicketPosition :one
-- move の before/after 指定チケットが現役かを確かめる（別スペース・アーカイブ済み・
-- 非実在はすべて 0 行に畳まれ、usecase は同じ拒否として扱う）。
SELECT "position" FROM tickets
WHERE workspace_id = $1 AND space_id = $2 AND id = $3 AND archived_at IS NULL AND deleted_at IS NULL;

-- =============================================================================
-- ticket_ranks（段 2: 並び順の正本。設計 Ⅳ-F）
-- =============================================================================

-- name: InsertTicketRank :exec
-- CreateTicket 成功直後に usecase が呼ぶ（tickets への INSERT とは別文。設計 Ⅳ-B の
-- 「tickets への INSERT は 1 本だけ」という縛りは ticket_ranks には及ばない）。
INSERT INTO ticket_ranks (workspace_id, ticket_id, context_kind, context_id, "position", created_at, updated_at)
VALUES ($1, $2, 'backlog', '00000000-0000-0000-0000-000000000000', $3, now(), now());

-- name: MoveTicketRank :execrows
UPDATE ticket_ranks
SET "position" = $3, updated_at = now()
WHERE workspace_id = $1 AND ticket_id = $2 AND context_kind = 'backlog';

-- name: LastActiveTicketRankPosition :one
-- ticket_ranks 自体は space_id を持たないので、対象スペースへの絞り込みは tickets への
-- JOIN で行う（LastActiveTicketPosition の ticket_ranks 版）。
SELECT COALESCE(max(r."position"), '')::text AS "position"
FROM ticket_ranks r
JOIN tickets t ON t.workspace_id = r.workspace_id AND t.id = r.ticket_id
WHERE r.workspace_id = $1 AND t.space_id = $2 AND r.context_kind = 'backlog'
  AND t.archived_at IS NULL AND t.deleted_at IS NULL;

-- =============================================================================
-- ticket_assignments
-- =============================================================================

-- name: UpsertTicketAssignment :one
-- 衝突キー (ticket_id) は PK 単独だが、行の所有者列 workspace_id が EXCLUDED と一致する
-- ときだけ更新する（blocks の upsert と同じ形。呼び出し側が ticket_id と workspace_id を
-- 取り違えても、他テナントの行を上書きしない歯止めになる）。
INSERT INTO ticket_assignments
  (workspace_id, ticket_id, assignee_principal_id, assigned_by_user_id, created_at)
VALUES ($1, $2, $3, $4, now())
ON CONFLICT (ticket_id)
DO UPDATE SET assignee_principal_id = EXCLUDED.assignee_principal_id,
              assigned_by_user_id = EXCLUDED.assigned_by_user_id
WHERE ticket_assignments.workspace_id = EXCLUDED.workspace_id
RETURNING *;

-- name: DeleteTicketAssignment :execrows
DELETE FROM ticket_assignments
WHERE workspace_id = $1 AND ticket_id = $2;

-- name: GetTicketAssignment :one
SELECT * FROM ticket_assignments
WHERE workspace_id = $1 AND ticket_id = $2;

-- name: ListTicketsAssignedToPrincipal :many
SELECT t.* FROM tickets t
JOIN ticket_assignments a ON a.workspace_id = t.workspace_id AND a.ticket_id = t.id
WHERE t.workspace_id = $1 AND a.assignee_principal_id = $2
  AND t.archived_at IS NULL AND t.deleted_at IS NULL;

-- =============================================================================
-- ticket_change_groups / ticket_change_items
-- =============================================================================

-- name: InsertTicketChangeGroup :one
INSERT INTO ticket_change_groups (id, workspace_id, ticket_id, actor_user_id, created_at)
VALUES ($1, $2, $3, $4, now())
RETURNING *;

-- name: InsertTicketChangeItem :exec
INSERT INTO ticket_change_items
  (id, workspace_id, group_id, field, old_value, new_value, old_label, new_label)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: ListTicketChangeGroups :many
SELECT * FROM ticket_change_groups
WHERE workspace_id = $1 AND ticket_id = $2
ORDER BY created_at DESC;

-- name: ListTicketChangeItemsByGroupIDs :many
-- group_id 群は json 配列 1 個のパラメータで渡す（no-array-param。comment.sql と同じ形）。
SELECT i.* FROM ticket_change_items i
WHERE i.workspace_id = sqlc.arg(workspace_id)
  AND i.group_id IN (
    SELECT value::uuid FROM json_array_elements_text(sqlc.arg(group_ids)::json) AS t(value)
  )
ORDER BY i.group_id;

-- =============================================================================
-- ticket_page_links / ticket_ticket_links（派生表）
-- =============================================================================

-- name: DeleteTicketPageLinksBySource :exec
DELETE FROM ticket_page_links
WHERE workspace_id = $1 AND source_ticket_id = $2;

-- name: InsertTicketPageLink :exec
INSERT INTO ticket_page_links (workspace_id, source_ticket_id, target_page_id)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: DeleteTicketTicketLinksBySource :exec
DELETE FROM ticket_ticket_links
WHERE workspace_id = $1 AND source_ticket_id = $2;

-- name: InsertTicketTicketLink :exec
INSERT INTO ticket_ticket_links (workspace_id, source_ticket_id, target_ticket_id)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: ListExistingPageIDsInWorkspace :many
-- 本文に貼られた pageRef 候補のうち、**そのワークスペースに実在するページ** だけを返す
-- （リンク切れ・別ワークスペースの ID は黙って除外する。設計 Ⅳ-I）。
SELECT id FROM pages
WHERE workspace_id = sqlc.arg(workspace_id)
  AND id IN (
    SELECT value::uuid FROM json_array_elements_text(sqlc.arg(page_ids)::json) AS t(value)
  );

-- name: ListExistingTicketIDsInWorkspace :many
SELECT id FROM tickets
WHERE workspace_id = sqlc.arg(workspace_id)
  AND id IN (
    SELECT value::uuid FROM json_array_elements_text(sqlc.arg(ticket_ids)::json) AS t(value)
  );

-- name: ListTicketPageLinksBySource :many
SELECT * FROM ticket_page_links
WHERE workspace_id = $1 AND source_ticket_id = $2;

-- name: ListPagesReferencingTicket :many
-- ticket-backlinks API の逆方向（そのページを参照しているチケット一覧）。
-- deleted_at IS NULL: 参照元チケットが削除されていれば、削除時に DeleteTicketPageLinksBySourceCascade
-- がこの行にも deleted_at を立てている。ここで除かないと、消えたはずのチケットの存在が
-- ページ側の逆参照一覧から漏れる（参照元チケットの deleted_at を都度 JOIN で見る代わりに、
-- 削除時に伝播させて 1 列で判定できるようにしてある。設計 Ⅳ-J の「読み出しの述語を単純に保つ」）。
SELECT * FROM ticket_page_links
WHERE workspace_id = $1 AND target_page_id = $2 AND deleted_at IS NULL;

-- name: ListTicketTicketLinksBySource :many
SELECT * FROM ticket_ticket_links
WHERE workspace_id = $1 AND source_ticket_id = $2;

-- name: ListTicketsReferencingTicket :many
-- deleted_at IS NULL: ListPagesReferencingTicket と同じ理由（参照元チケットの削除を伝播で判定）。
SELECT * FROM ticket_ticket_links
WHERE workspace_id = $1 AND target_ticket_id = $2 AND deleted_at IS NULL;

-- name: GetTicketAcrossWorkspaces :one
-- チケットを **ID だけ** で引く。/kb/tickets/{ticketId} の URL からワークスペースを
-- 特定するための、このファイルで唯一 workspace_id を WHERE に持たない読み取り
-- （knowledge_base.sql の GetPageAcrossWorkspaces と同じ役割・同じ作法）。
-- 引いた直後に必ずその workspace の権限判定を通すこと（判定なしで応答に使わない）。
-- id は uuid の主キーで全テナント一意なので、これ自体が越境にはならない。
SELECT id, workspace_id, space_id FROM tickets
WHERE id = $1;

-- name: CountActiveTicketsGroupedByStatus :many
-- 管理画面の「使用中 N 件」。状態 1 つずつ CountActiveTicketsByStatus を呼ぶと
-- 状態の数だけ問い合わせが増えるので、スペース 1 回の GROUP BY でまとめて数える。
-- 現役（archived_at IS NULL）だけを数えるのは、アーカイブ済みのチケットが
-- 状態のアーカイブを妨げないため（usecase の 409 判定と同じ範囲に揃える）。
SELECT status_id, count(*)::bigint AS count FROM tickets
WHERE workspace_id = $1 AND space_id = $2 AND archived_at IS NULL
GROUP BY status_id;

-- name: CountActiveTicketsGroupedByType :many
SELECT type_id, count(*)::bigint AS count FROM tickets
WHERE workspace_id = $1 AND space_id = $2 AND archived_at IS NULL
GROUP BY type_id;
