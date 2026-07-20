import apiClient from '@/shared/api/axios';
import { ADMIN } from '@/shared/config/apiRoutes';

/** 監査イベント（管理者の重要操作の記録）。super_admin 専用。 */
export interface AuditEvent {
  id: number;
  actorId: number;
  actorEmail: string;
  actorRole: string;
  /** 「METHOD ルートパターン」（例: "PATCH /api/v2/admin/companies/:id/active"）。 */
  action: string;
  targetId: number;
  createdAt: string;
}

export const AuditRepository = {
  /** 監査ログを新しい順で取得する（super_admin 専用）。 */
  async list(): Promise<AuditEvent[]> {
    const { data } = await apiClient.get<AuditEvent[]>(ADMIN.auditEvents);
    return data;
  },
};
