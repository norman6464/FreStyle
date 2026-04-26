import apiClient from '../lib/axios';

export interface AdminScenario {
  id: number;
  name: string;
  description: string | null;
  category: string | null;
  roleName: string;
  difficulty: string | null;
  systemPrompt: string;
}

export interface AdminScenarioForm {
  name: string;
  description?: string;
  category?: string;
  roleName: string;
  difficulty?: string;
  systemPrompt: string;
}

/**
 * 管理者専用: 練習シナリオ CRUD。
 * 403 が返る場合は admin 権限なし（Cognito の admin グループ未所属）。
 */
class AdminScenarioRepository {
  async list(): Promise<AdminScenario[]> {
    const res = await apiClient.get<AdminScenario[]>('/api/admin/scenarios');
    return res.data;
  }

  async create(form: AdminScenarioForm): Promise<AdminScenario> {
    const res = await apiClient.post<AdminScenario>('/api/admin/scenarios', form);
    return res.data;
  }

  async update(id: number, form: AdminScenarioForm): Promise<AdminScenario> {
    const res = await apiClient.put<AdminScenario>(`/api/admin/scenarios/${id}`, form);
    return res.data;
  }

  async remove(id: number): Promise<void> {
    await apiClient.delete(`/api/admin/scenarios/${id}`);
  }
}

export default new AdminScenarioRepository();
