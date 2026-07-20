import api from '@/shared/api/axios';
import { DASHBOARD, CHAPTER_VIEW } from '@/shared/config/apiRoutes';
import type { UserDashboard } from '../types';

/** ダッシュボード API ラッパ。 */
const DashboardRepository = {
  async get(): Promise<UserDashboard> {
    const res = await api.get<UserDashboard>(DASHBOARD.get);
    return res.data;
  },

  /** 章閲覧記録（ベストエフォート — エラーは握り潰す）。 */
  async recordChapterView(teachingMaterialId: number): Promise<void> {
    try {
      await api.post(CHAPTER_VIEW.record(teachingMaterialId));
    } catch {
      // ベストエフォート：失敗してもユーザー体験に影響しない
    }
  },
};

export default DashboardRepository;
