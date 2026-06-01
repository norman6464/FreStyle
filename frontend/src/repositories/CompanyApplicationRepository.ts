import apiClient from '../lib/axios';
import { COMPANY_APPLICATIONS } from '../constants/apiRoutes';

/** 企業利用申請フォームの送信内容（公開・認証不要）。 */
export interface CompanyApplicationForm {
  companyName: string;
  applicantName: string;
  email: string;
  message: string;
}

export const CompanyApplicationRepository = {
  /** 企業利用申請を送信する。認証不要。 */
  async apply(form: CompanyApplicationForm): Promise<void> {
    await apiClient.post(COMPANY_APPLICATIONS.create, form);
  },
};
