import api from '@/shared/api/axios';
import { ADMIN } from '@/shared/config/apiRoutes';

/** 直近アクティブメンバー 1 人分（backend usecase.MemberLearningSummaryItem と 1:1）。 */
export interface MemberLearningSummaryItem {
  userId: number;
  name: string;
  /** 最後に学習活動があった日（YYYY-MM-DD、UTC 基準）。 */
  lastActiveDate: string;
  /** 直近 7 日間の活動回数合計。 */
  recentActivityCount: number;
}

/** 自社メンバーの学習状況サマリー（backend usecase.CompanyLearningSummaryOutput と 1:1）。 */
export interface CompanyLearningSummary {
  traineeCount: number;
  activeToday: number;
  activeThisWeek: number;
  recentMembers: MemberLearningSummaryItem[];
}

/** 従業員一覧の 1 行（backend handler.memberResponse と 1:1）。 */
export interface Member {
  id: number;
  email: string;
  displayName: string;
  role: string;
  /** AI 利用可否の個別上書き。null = 会社設定に従う。 */
  aiChatEnabled: boolean | null;
  /** アカウントの有効/無効。false = 無効（ログイン/利用不可）。 */
  isActive: boolean;
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

  /** 従業員アカウントの有効/無効を切り替える（false で停止 → ログイン/利用不可）。 */
  async updateActive(userId: number, active: boolean): Promise<void> {
    await api.patch(ADMIN.memberActive(userId), { active });
  },

  /** 従業員を論理削除する（一覧から退会させる）。 */
  async remove(userId: number): Promise<void> {
    await api.delete(ADMIN.member(userId));
  },

  /** 自社メンバーの学習状況サマリーを取得する（company_admin のホーム用）。 */
  async learningSummary(): Promise<CompanyLearningSummary> {
    const res = await api.get<CompanyLearningSummary>(ADMIN.membersLearningSummary);
    return res.data;
  },
};

export default AdminMemberRepository;
