import apiClient from '../lib/axios';
import { COMPANY_APPLICATIONS } from '../constants/apiRoutes';

/** 企業利用申請フォームの送信内容（公開・認証不要）。 */
export interface CompanyApplicationForm {
  companyName: string;
  applicantName: string;
  email: string;
  message: string;
}

/** 企業利用申請のステータス（backend の domain と一致）。 */
export type CompanyApplicationStatus = 'pending' | 'approved' | 'rejected';

/** 企業利用申請（super_admin が一覧で確認し、承認 / 却下する）。 */
export interface CompanyApplication {
  id: number;
  companyName: string;
  applicantName: string;
  email: string;
  message: string;
  status: CompanyApplicationStatus;
  createdAt: string;
  updatedAt: string;
}

export const CompanyApplicationRepository = {
  /** 企業利用申請を送信する。認証不要。 */
  async apply(form: CompanyApplicationForm): Promise<void> {
    await apiClient.post(COMPANY_APPLICATIONS.create, form);
  },

  /** super_admin: 全申請を新しい順で取得する。 */
  async adminList(): Promise<CompanyApplication[]> {
    const { data } = await apiClient.get<CompanyApplication[]>(COMPANY_APPLICATIONS.adminList);
    return data;
  },

  /** super_admin: 申請の status を更新する（approved / rejected / pending）。 */
  async adminUpdateStatus(id: number, status: CompanyApplicationStatus): Promise<void> {
    await apiClient.patch(COMPANY_APPLICATIONS.adminUpdateStatus(id), { status });
  },
};
