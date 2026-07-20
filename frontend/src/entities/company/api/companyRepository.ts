import apiClient from '@/shared/api/axios';
import { ADMIN } from '@/shared/config/apiRoutes';

export interface Company {
  id: number;
  name: string;
  /** 会社アカウントの有効/無効。false = 無効（その会社の全ユーザーが利用不可）。 */
  isActive: boolean;
  createdAt: string;
  updatedAt: string;
}

/** 会社横断ビューの 1 行（会社 + メンバー集計）。super_admin 専用。 */
export interface CompanyStat {
  id: number;
  name: string;
  isActive: boolean;
  createdAt: string;
  /** 在籍メンバー総数（論理削除・会社未所属は除外）。 */
  memberTotal: number;
  /** 有効（is_active）なメンバー数。 */
  activeMembers: number;
  /** trainee（受講者）のメンバー数。 */
  traineeCount: number;
}

class CompanyRepository {
  async list(): Promise<Company[]> {
    const res = await apiClient.get<Company[]>(ADMIN.companies);
    return res.data;
  }

  /** 各社のメンバー集計付きの会社横断ビューを取得する（super_admin 専用）。 */
  async listStats(): Promise<CompanyStat[]> {
    const res = await apiClient.get<CompanyStat[]>(ADMIN.companiesStats);
    return res.data;
  }

  /** 会社アカウントの有効/無効を切り替える（super_admin 専用）。 */
  async updateActive(id: number, active: boolean): Promise<void> {
    await apiClient.patch(ADMIN.companyActive(id), { active });
  }
}

export default new CompanyRepository();
