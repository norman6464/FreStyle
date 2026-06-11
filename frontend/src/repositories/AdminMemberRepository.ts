import api from '../lib/axios';
import { ADMIN } from '../constants/apiRoutes';

/** 従業員一覧の 1 行（backend handler.memberResponse と 1:1）。 */
export interface Member {
  id: number;
  email: string;
  displayName: string;
  role: string;
  /** AI 利用可否の個別上書き。null = 会社設定に従う。 */
  aiChatEnabled: boolean | null;
}

/**
 * company_admin / super_admin 向けの従業員管理 API ラッパー。
 * 自社の従業員一覧取得と、各従業員の AI 利用可否の個別更新を扱う。
 */
const AdminMemberRepository = {
  async listMembers(): Promise<Member[]> {
    const res = await api.get<Member[]>(ADMIN.members);
    return res.data;
  },

  /**
   * 従業員の AI 利用可否を個別更新する。enabled=null で会社設定に従う状態へ戻す。
   */
  async updateAiAccess(userId: number, enabled: boolean | null): Promise<void> {
    await api.patch(ADMIN.memberAiAccess(userId), { enabled });
  },
};

export default AdminMemberRepository;
