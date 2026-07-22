import api from '@/shared/api/axios';
import { LESSON_PROGRESS } from '@/shared/config/apiRoutes';
import type { UserLessonProgress } from '../model/types';

/**
 * 学習進捗 API ラッパ。
 *
 * すべて current user 名義（backend が context から user を解決）。 userId は渡さない（IDOR 対策）。
 * course は教材から backend が解決するため、 フロントは teachingMaterialId だけ渡す。
 */
const LessonProgressRepository = {
  async list(): Promise<UserLessonProgress[]> {
    const res = await api.get<UserLessonProgress[]>(LESSON_PROGRESS.list);
    return res.data;
  },

  async complete(teachingMaterialId: number): Promise<void> {
    await api.post(LESSON_PROGRESS.complete, { teachingMaterialId });
  },

  async incomplete(teachingMaterialId: number): Promise<void> {
    await api.delete(LESSON_PROGRESS.incomplete(teachingMaterialId));
  },
};

export default LessonProgressRepository;
