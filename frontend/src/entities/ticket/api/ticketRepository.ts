import apiClient from '@/shared/api/axios';
import { TICKET_API } from '@/shared/config/apiRoutes';
import { toArray } from '@/shared/lib/toArray';
import type {
  EnableTicketsResult,
  Label,
  ResolvedTicket,
  Ticket,
  TicketAssignment,
  TicketChangeGroup,
  TicketHierarchyLevel,
  TicketListFilter,
  TicketPriority,
  TicketResolution,
  TicketStatus,
  TicketStatusCategory,
  TicketStatusWire,
  TicketType,
  TicketTypeWire,
  TicketWire,
} from '../model/types';

/**
 * チケット・バックログの repository。`entities/kb/api/kbRepository.ts` と同じ形
 * （プレーンオブジェクト + default export、axios は `@/shared/api/axios` の 1 個だけを使う
 * ので認証ヘッダは何もしなくてよい）。
 *
 * 失敗は例外として投げる。ここで握り潰して null / false を返すと、呼び出し側は
 * 失敗を知りようがない。
 */

function normalizeTicket(wire: TicketWire): Ticket {
  return {
    id: wire.id,
    workspaceId: wire.workspaceId,
    spaceId: wire.spaceId,
    number: wire.number,
    typeId: wire.typeId,
    statusId: wire.statusId,
    parentId: wire.parentId ?? null,
    title: wire.title,
    doc: wire.doc,
    priority: wire.priority,
    startDate: wire.startDate ?? null,
    dueDate: wire.dueDate ?? null,
    position: wire.position,
    closedAt: wire.closedAt ?? null,
    resolution: wire.resolution ?? null,
    createdByUserId: wire.createdByUserId,
    archivedAt: wire.archivedAt ?? null,
    createdAt: wire.createdAt,
    updatedAt: wire.updatedAt,
    assigneePrincipalId: wire.assigneePrincipalId ?? null,
    labels: toArray<Label>(wire.labels),
  };
}

// 状態・種別の Create/Update 応答は activeTicketCount を持たない（backend が domain 構造体を
// そのまま返すだけの経路のため）。呼び出し側（useTicketMasters）はこの正規化のあとに
// 必ず一覧を取り直すので、ここでの 0 は一瞬しか見えない暫定値でよい（設計 Ⅶ「状態 / 種別の
// 変更 → マスタを取り直す」）。
function normalizeTicketStatus(wire: TicketStatusWire): TicketStatus {
  return {
    id: wire.id,
    workspaceId: wire.workspaceId,
    spaceId: wire.spaceId,
    name: wire.name,
    category: wire.category,
    color: wire.color,
    position: wire.position,
    isInitial: wire.isInitial,
    archivedAt: wire.archivedAt ?? null,
    createdAt: wire.createdAt,
    updatedAt: wire.updatedAt,
    activeTicketCount: wire.activeTicketCount ?? 0,
  };
}

function normalizeTicketType(wire: TicketTypeWire): TicketType {
  return {
    id: wire.id,
    workspaceId: wire.workspaceId,
    spaceId: wire.spaceId,
    name: wire.name,
    hierarchyLevel: wire.hierarchyLevel,
    color: wire.color,
    position: wire.position,
    isDefault: wire.isDefault,
    templateTitle: wire.templateTitle ?? null,
    templateDoc: wire.templateDoc ?? null,
    archivedAt: wire.archivedAt ?? null,
    createdAt: wire.createdAt,
    updatedAt: wire.updatedAt,
    activeTicketCount: wire.activeTicketCount ?? 0,
  };
}

export interface CreateTicketInput {
  parentId?: string;
  typeId?: string;
  statusId?: string;
  title: string;
  doc?: unknown;
  priority?: TicketPriority;
  startDate?: string;
  dueDate?: string;
}

export interface UpdateTicketInput {
  title: string;
  doc: unknown;
  typeId: string;
  priority: TicketPriority;
  startDate?: string | null;
  dueDate?: string | null;
}

export interface MoveTicketInput {
  anchorTicketId?: string;
  anchorAfter?: boolean;
}

export interface ChangeTicketStatusInput {
  statusId: string;
  resolution?: TicketResolution;
}

export interface TicketStatusInput {
  name: string;
  category: TicketStatusCategory;
  color: string;
}

export interface TicketTypeInput {
  name: string;
  hierarchyLevel: TicketHierarchyLevel;
  color: string;
}

const TicketRepository = {
  async enable(
    workspaceSlug: string,
    spaceId: string,
    sourceSpaceId?: string,
  ): Promise<EnableTicketsResult> {
    const res = await apiClient.post<EnableTicketsResult>(TICKET_API.enable(workspaceSlug, spaceId), {
      sourceSpaceId: sourceSpaceId ?? '',
    });
    return res.data;
  },

  async fetchTickets(
    workspaceSlug: string,
    spaceId: string,
    filter: TicketListFilter = {},
  ): Promise<Ticket[]> {
    const params: Record<string, string> = {};
    if (filter.statusId) params.statusId = filter.statusId;
    if (filter.typeId) params.typeId = filter.typeId;
    if (filter.assigneePrincipalId) params.assigneePrincipalId = filter.assigneePrincipalId;
    if (filter.archived) params.archived = 'true';
    const res = await apiClient.get<{ tickets: TicketWire[] }>(TICKET_API.tickets(workspaceSlug, spaceId), {
      params,
    });
    return toArray<TicketWire>(res.data?.tickets).map(normalizeTicket);
  },

  async createTicket(workspaceSlug: string, spaceId: string, input: CreateTicketInput): Promise<Ticket> {
    const res = await apiClient.post<TicketWire>(TICKET_API.tickets(workspaceSlug, spaceId), {
      parentId: input.parentId ?? '',
      typeId: input.typeId ?? '',
      statusId: input.statusId ?? '',
      title: input.title,
      doc: input.doc,
      priority: input.priority ?? 0,
      startDate: input.startDate,
      dueDate: input.dueDate,
    });
    return normalizeTicket(res.data);
  },

  async fetchTicket(workspaceSlug: string, ticketId: string): Promise<Ticket> {
    const res = await apiClient.get<TicketWire>(TICKET_API.ticket(workspaceSlug, ticketId));
    return normalizeTicket(res.data);
  },

  async fetchTicketByKey(workspaceSlug: string, key: string): Promise<Ticket> {
    const res = await apiClient.get<TicketWire>(TICKET_API.ticketByKey(workspaceSlug, key));
    return normalizeTicket(res.data);
  },

  async resolveTicket(ticketId: string): Promise<ResolvedTicket> {
    const res = await apiClient.get<{
      workspaceSlug: string;
      workspaceName: string;
      ticket: TicketWire;
      canEdit: boolean;
    }>(TICKET_API.resolveTicket(ticketId));
    return {
      workspaceSlug: res.data.workspaceSlug,
      workspaceName: res.data.workspaceName,
      ticket: normalizeTicket(res.data.ticket),
      canEdit: res.data.canEdit,
    };
  },

  async updateTicket(workspaceSlug: string, ticketId: string, input: UpdateTicketInput): Promise<Ticket> {
    const res = await apiClient.put<TicketWire>(TICKET_API.ticket(workspaceSlug, ticketId), {
      title: input.title,
      doc: input.doc,
      typeId: input.typeId,
      priority: input.priority,
      startDate: input.startDate ?? undefined,
      dueDate: input.dueDate ?? undefined,
    });
    return normalizeTicket(res.data);
  },

  /** 204 応答。並び替え後の順位は一覧の取り直しでしか分からない（設計 Ⅶ）。 */
  async moveTicket(workspaceSlug: string, ticketId: string, input: MoveTicketInput): Promise<void> {
    await apiClient.post(TICKET_API.moveTicket(workspaceSlug, ticketId), {
      anchorTicketId: input.anchorTicketId ?? '',
      anchorAfter: input.anchorAfter ?? false,
    });
  },

  async archiveTicket(workspaceSlug: string, ticketId: string): Promise<Ticket> {
    const res = await apiClient.post<TicketWire>(TICKET_API.archiveTicket(workspaceSlug, ticketId));
    return normalizeTicket(res.data);
  },

  async restoreTicket(workspaceSlug: string, ticketId: string): Promise<Ticket> {
    const res = await apiClient.post<TicketWire>(TICKET_API.restoreTicket(workspaceSlug, ticketId));
    return normalizeTicket(res.data);
  },

  async changeTicketStatus(
    workspaceSlug: string,
    ticketId: string,
    input: ChangeTicketStatusInput,
  ): Promise<Ticket> {
    const res = await apiClient.post<TicketWire>(TICKET_API.changeTicketStatus(workspaceSlug, ticketId), {
      statusId: input.statusId,
      resolution: input.resolution,
    });
    return normalizeTicket(res.data);
  },

  /** parentId に null（またはキー省略）を渡すとトップレベルへ戻す。 */
  async changeTicketParent(workspaceSlug: string, ticketId: string, parentId: string | null): Promise<Ticket> {
    const res = await apiClient.put<TicketWire>(TICKET_API.changeTicketParent(workspaceSlug, ticketId), {
      parentId: parentId ?? '',
    });
    return normalizeTicket(res.data);
  },

  async assignTicket(
    workspaceSlug: string,
    ticketId: string,
    assigneePrincipalId: string,
  ): Promise<TicketAssignment> {
    const res = await apiClient.put<TicketAssignment>(TICKET_API.ticketAssignee(workspaceSlug, ticketId), {
      assigneePrincipalId,
    });
    return res.data;
  },

  /** 204 応答。 */
  async unassignTicket(workspaceSlug: string, ticketId: string): Promise<void> {
    await apiClient.delete(TICKET_API.ticketAssignee(workspaceSlug, ticketId));
  },

  async fetchTicketHistory(workspaceSlug: string, ticketId: string): Promise<TicketChangeGroup[]> {
    const res = await apiClient.get<{ groups: TicketChangeGroup[] }>(
      TICKET_API.ticketHistory(workspaceSlug, ticketId),
    );
    return toArray<TicketChangeGroup>(res.data?.groups);
  },

  async fetchTicketStatuses(
    workspaceSlug: string,
    spaceId: string,
    archived = false,
  ): Promise<TicketStatus[]> {
    const res = await apiClient.get<{ statuses: TicketStatusWire[] }>(
      TICKET_API.ticketStatuses(workspaceSlug, spaceId),
      { params: archived ? { archived: 'true' } : undefined },
    );
    return toArray<TicketStatusWire>(res.data?.statuses).map(normalizeTicketStatus);
  },

  async createTicketStatus(workspaceSlug: string, spaceId: string, input: TicketStatusInput): Promise<TicketStatus> {
    const res = await apiClient.post<TicketStatusWire>(TICKET_API.ticketStatuses(workspaceSlug, spaceId), input);
    return normalizeTicketStatus(res.data);
  },

  async updateTicketStatus(
    workspaceSlug: string,
    spaceId: string,
    statusId: string,
    input: TicketStatusInput,
  ): Promise<TicketStatus> {
    const res = await apiClient.put<TicketStatusWire>(
      TICKET_API.ticketStatus(workspaceSlug, spaceId, statusId),
      input,
    );
    return normalizeTicketStatus(res.data);
  },

  /** 204 応答。 */
  async setInitialTicketStatus(workspaceSlug: string, spaceId: string, statusId: string): Promise<void> {
    await apiClient.post(TICKET_API.setInitialTicketStatus(workspaceSlug, spaceId, statusId));
  },

  /** 204 応答。使用中は 409 status_in_use。 */
  async archiveTicketStatus(workspaceSlug: string, spaceId: string, statusId: string): Promise<void> {
    await apiClient.post(TICKET_API.archiveTicketStatus(workspaceSlug, spaceId, statusId));
  },

  /** 204 応答。 */
  async restoreTicketStatus(workspaceSlug: string, spaceId: string, statusId: string): Promise<void> {
    await apiClient.post(TICKET_API.restoreTicketStatus(workspaceSlug, spaceId, statusId));
  },

  async fetchTicketTypes(workspaceSlug: string, spaceId: string, archived = false): Promise<TicketType[]> {
    const res = await apiClient.get<{ types: TicketTypeWire[] }>(TICKET_API.ticketTypes(workspaceSlug, spaceId), {
      params: archived ? { archived: 'true' } : undefined,
    });
    return toArray<TicketTypeWire>(res.data?.types).map(normalizeTicketType);
  },

  async createTicketType(workspaceSlug: string, spaceId: string, input: TicketTypeInput): Promise<TicketType> {
    const res = await apiClient.post<TicketTypeWire>(TICKET_API.ticketTypes(workspaceSlug, spaceId), input);
    return normalizeTicketType(res.data);
  },

  async updateTicketType(
    workspaceSlug: string,
    spaceId: string,
    typeId: string,
    input: TicketTypeInput,
  ): Promise<TicketType> {
    const res = await apiClient.put<TicketTypeWire>(TICKET_API.ticketType(workspaceSlug, spaceId, typeId), input);
    return normalizeTicketType(res.data);
  },

  /** 204 応答。 */
  async setDefaultTicketType(workspaceSlug: string, spaceId: string, typeId: string): Promise<void> {
    await apiClient.post(TICKET_API.setDefaultTicketType(workspaceSlug, spaceId, typeId));
  },

  /** 204 応答。使用中は 409 type_in_use。 */
  async archiveTicketType(workspaceSlug: string, spaceId: string, typeId: string): Promise<void> {
    await apiClient.post(TICKET_API.archiveTicketType(workspaceSlug, spaceId, typeId));
  },

  /** 204 応答。 */
  async restoreTicketType(workspaceSlug: string, spaceId: string, typeId: string): Promise<void> {
    await apiClient.post(TICKET_API.restoreTicketType(workspaceSlug, spaceId, typeId));
  },
};

export default TicketRepository;
