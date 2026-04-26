import apiClient from '../lib/axios';

export interface AdminInvitation {
  id: number;
  companyId: number;
  email: string;
  role: 'trainee' | 'company_admin';
  invitedBy: number | null;
  expiresAt: string;
  acceptedAt: string | null;
  acceptedUserId: number | null;
  createdAt: string;
}

export interface CreateInvitationForm {
  email: string;
  role: 'trainee' | 'company_admin';
  displayName?: string;
}

/** 管理者専用: メール招待 CRUD */
class AdminInvitationRepository {
  async list(): Promise<AdminInvitation[]> {
    const res = await apiClient.get<AdminInvitation[]>('/api/admin/invitations');
    return res.data;
  }

  async create(form: CreateInvitationForm): Promise<AdminInvitation> {
    const res = await apiClient.post<AdminInvitation>('/api/admin/invitations', form);
    return res.data;
  }

  async cancel(id: number): Promise<void> {
    await apiClient.delete(`/api/admin/invitations/${id}`);
  }
}

export default new AdminInvitationRepository();
