import api from '@/shared/api/axios';
import { CHAPTER_VIEW } from '@/shared/config/apiRoutes';

/** 章閲覧記録 API ラッパ。 */
const ChapterViewRepository = {
  /** 章閲覧記録（ベストエフォート — エラーは握り潰す）。 */
  async recordChapterView(teachingMaterialId: number): Promise<void> {
    try {
      await api.post(CHAPTER_VIEW.record(teachingMaterialId));
    } catch {
      // ベストエフォート：失敗してもユーザー体験に影響しない
    }
  },
};

export default ChapterViewRepository;
