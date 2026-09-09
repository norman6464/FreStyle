/**
 * チケット・バックログのドメイン型。backend の応答（`ticketResponse` 系。
 * `backend/internal/handler/ticket_handler.go` `ticket_status_handler.go` `ticket_type_handler.go`
 * 参照）と 1:1 対応させる。
 *
 * Go 側の `omitempty` フィールド（parentId / startDate / dueDate / closedAt / resolution /
 * archivedAt / assigneePrincipalId）は値が無いとキー自体が応答から消える。ここでは
 * `?: T | null` ではなく `?: T` のまま宣言し、`ticketRepository.ts` の normalize 関数が
 * `?? null` に畳んでから返す（entities/kb の TicketWire 相当のパターン）ので、
 * repository の外（hook / component）ではすべてのフィールドが `null` か値のどちらかで
 * 揃っている前提でよい。
 */

/** domain.TicketPriority（1=高 2=中(既定) 3=低）。 */
export type TicketPriority = 1 | 2 | 3;

/** domain.TicketStatusCategory。 */
export type TicketStatusCategory = 'todo' | 'in_progress' | 'done';

/** domain.TicketResolution。category = 'done' のときだけ意味を持つ。 */
export type TicketResolution = 'done' | 'wont_do' | 'invalid' | 'duplicate' | 'cannot_reproduce';

/** domain.TicketType.HierarchyLevel（1=束ね 0=標準 -1=小作業）。 */
export type TicketHierarchyLevel = 1 | 0 | -1;

/** domain.TicketChangeField。 */
export type TicketChangeField =
  | 'title'
  | 'doc'
  | 'status'
  | 'type'
  | 'priority'
  | 'assignee'
  | 'parent'
  | 'start_date'
  | 'due_date'
  | 'resolution'
  | 'position'
  | 'archived'
  | 'category'
  | 'milestone'
  | 'link';

/** backend の wire 形（ticketRepository.ts の内部でのみ使う。外へは Ticket として出す）。 */
export interface TicketWire {
  id: string;
  workspaceId: string;
  spaceId: string;
  number: number;
  typeId: string;
  statusId: string;
  parentId?: string;
  title: string;
  doc: unknown;
  priority: TicketPriority;
  startDate?: string;
  dueDate?: string;
  position: string;
  closedAt?: string;
  resolution?: TicketResolution;
  createdByUserId: number;
  archivedAt?: string;
  createdAt: string;
  updatedAt: string;
  assigneePrincipalId?: string;
}

/** normalizeTicket 後の形。すべてのフィールドが揃っている（省略は無い）。 */
export interface Ticket {
  id: string;
  workspaceId: string;
  spaceId: string;
  number: number;
  typeId: string;
  statusId: string;
  parentId: string | null;
  title: string;
  doc: unknown;
  priority: TicketPriority;
  startDate: string | null;
  dueDate: string | null;
  position: string;
  closedAt: string | null;
  resolution: TicketResolution | null;
  createdByUserId: number;
  archivedAt: string | null;
  createdAt: string;
  updatedAt: string;
  assigneePrincipalId: string | null;
}

/** チケットの表示キー（例 FRESTYLE-12）。spaceKey + number から組み立てる（lib/ticketKey.ts）。 */
export type TicketKey = string;

export interface TicketStatusWire {
  id: string;
  workspaceId: string;
  spaceId: string;
  name: string;
  category: TicketStatusCategory;
  color: string;
  position: string;
  isInitial: boolean;
  archivedAt?: string;
  createdAt: string;
  updatedAt: string;
  activeTicketCount: number;
}

export interface TicketStatus {
  id: string;
  workspaceId: string;
  spaceId: string;
  name: string;
  category: TicketStatusCategory;
  color: string;
  position: string;
  isInitial: boolean;
  archivedAt: string | null;
  createdAt: string;
  updatedAt: string;
  /** 現役チケットでの使用数（管理画面の「使用中 N 件」）。 */
  activeTicketCount: number;
}

export interface TicketTypeWire {
  id: string;
  workspaceId: string;
  spaceId: string;
  name: string;
  hierarchyLevel: TicketHierarchyLevel;
  color: string;
  position: string;
  isDefault: boolean;
  templateTitle?: string;
  templateDoc?: unknown;
  archivedAt?: string;
  createdAt: string;
  updatedAt: string;
  activeTicketCount: number;
}

export interface TicketType {
  id: string;
  workspaceId: string;
  spaceId: string;
  name: string;
  hierarchyLevel: TicketHierarchyLevel;
  color: string;
  position: string;
  isDefault: boolean;
  templateTitle: string | null;
  templateDoc: unknown | null;
  archivedAt: string | null;
  createdAt: string;
  updatedAt: string;
  activeTicketCount: number;
}

/** PUT .../tickets/:id/assignee の応答（domain.TicketAssignment）。 */
export interface TicketAssignment {
  workspaceId: string;
  ticketId: string;
  assigneePrincipalId: string;
  assignedByUserId: number;
  createdAt: string;
}

export interface TicketChangeItem {
  id: string;
  groupId: string;
  field: TicketChangeField;
  oldValue: string | null;
  newValue: string | null;
  oldLabel: string | null;
  newLabel: string | null;
}

export interface TicketChangeGroup {
  id: string;
  workspaceId: string;
  ticketId: string;
  actorUserId: number;
  createdAt: string;
  items: TicketChangeItem[];
}

/** GET /api/v2/kb/tickets/:ticketId（slug 無し解決）の応答。 */
export interface ResolvedTicket {
  workspaceSlug: string;
  workspaceName: string;
  ticket: Ticket;
  canEdit: boolean;
}

/** チケット一覧の絞り込み（List のクエリパラメータに対応）。 */
export interface TicketListFilter {
  statusId?: string;
  typeId?: string;
  assigneePrincipalId?: string;
  /** true でアーカイブ済みだけを返す（現役との「込み」は取れない。設計 Ⅳ-C）。 */
  archived?: boolean;
}

/** POST .../tickets/enable の応答。 */
export interface EnableTicketsResult {
  statusCount: number;
  typeCount: number;
}
