import api from '@/shared/api/axios';
import { toArray } from '@/shared/lib/toArray';
import { ADMIN } from '@/shared/config/apiRoutes';

/** 従業員一覧の 1 行（backend handler.memberResponse と 1:1）。 */
export interface Member {
  id: number;
  email: string;
  displayName: string;
  role: string;
  /** アカウントの有効/無効。false = 無効（ログイン/利用不可）。 */
  isActive: boolean;
}

/**
 * company_admin / super_admin 向けの従業員管理 API ラッパー。
 * 自社の従業員一覧取得と、アカウントの有効/無効・削除を扱う。
 */
const AdminMemberRepository = {
  async listMembers(): Promise<Member[]> {
    const res = await api.get<Member[]>(ADMIN.members);
    return toArray<Member>(res.data);
  },

  /** 従業員アカウントの有効/無効を切り替える（false で停止 → ログイン/利用不可）。 */
  async updateActive(userId: number, active: boolean): Promise<void> {
    await api.patch(ADMIN.memberActive(userId), { active });
  },

  /** 従業員を論理削除する（一覧から退会させる）。 */
  async remove(userId: number): Promise<void> {
    await api.delete(ADMIN.member(userId));
  },
};

export default AdminMemberRepository;
