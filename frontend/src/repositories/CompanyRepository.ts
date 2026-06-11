import apiClient from '../lib/axios';
import { ADMIN } from '../constants/apiRoutes';

export interface Company {
  id: number;
  name: string;
  /** 会社アカウントの有効/無効。false = 無効（その会社の全ユーザーが利用不可）。 */
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

class CompanyRepository {
  async list(): Promise<Company[]> {
    const res = await apiClient.get<Company[]>(ADMIN.companies);
    return res.data;
  }

  /** 会社アカウントの有効/無効を切り替える（super_admin 専用）。 */
  async updateActive(id: number, active: boolean): Promise<void> {
    await apiClient.patch(ADMIN.companyActive(id), { active });
  }
}

export default new CompanyRepository();
