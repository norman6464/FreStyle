import apiClient from '@/shared/api/axios';
import { toArray } from '@/shared/lib/toArray';
import { ADMIN } from '@/shared/config/apiRoutes';

export interface AdminInvitation {
  id: number;
  workspaceId?: string;
  email: string;
  role: 'trainee' | 'company_admin';
  invitedBy: number | null;
  expiresAt: string;
  acceptedAt: string | null;
  acceptedUserId: number | null;
  createdAt: string;
}

export type InvitationMethod = 'magic_link' | 'temporary_password';

export interface CreateInvitationForm {
  email: string;
  role: 'trainee' | 'company_admin';
  displayName?: string;
  /** 招待方式。未指定はマジックリンク（従来どおりメール送信）。 */
  method?: InvitationMethod;
}

/** temporary_password 方式のレスポンス。temporaryPassword は 1 度だけ返る（保存不可）。 */
export interface TemporaryPasswordInvitation {
  invitation: AdminInvitation;
  temporaryPassword: string;
}

/** 管理者専用: メール招待 CRUD */
class AdminInvitationRepository {
  async list(): Promise<AdminInvitation[]> {
    const res = await apiClient.get<AdminInvitation[]>(ADMIN.invitations);
    return toArray<AdminInvitation>(res.data);
  }

  async create(form: CreateInvitationForm): Promise<AdminInvitation> {
    const res = await apiClient.post<AdminInvitation>(ADMIN.invitations, form);
    return res.data;
  }

  /**
   * 初期パスワード方式で招待する。Cognito 一時パスワードを発行し、
   * レスポンスで 1 度だけ受け取る（保存・再取得はできない）。
   */
  async createWithTemporaryPassword(
    form: CreateInvitationForm
  ): Promise<TemporaryPasswordInvitation> {
    const res = await apiClient.post<TemporaryPasswordInvitation>(ADMIN.invitations, {
      ...form,
      method: 'temporary_password',
    });
    return res.data;
  }

  async cancel(id: number): Promise<void> {
    await apiClient.delete(ADMIN.invitationById(id));
  }
}

export default new AdminInvitationRepository();
