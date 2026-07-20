import api from '@/shared/api/axios';
import { COURSES } from '@/shared/config/apiRoutes';
import type { Course, CourseWithProgress, TeachingMaterial, UserChapterView } from '../types';

/**
 * コース API ラッパ。
 *
 * actor の role / company は backend 側で context から取り出して自動フィルタするため、
 * フロントは companyId を渡さない（IDOR 対策）。
 */
export interface CoursePayload {
  title: string;
  description: string;
  /** 学習領域カテゴリ（空 = 未分類） */
  category: string;
  /** 主に扱う言語・技術（空 = 言語が主題でない） */
  language: string;
  sortOrder: number;
  isPublished: boolean;
}

const CourseRepository = {
  async list(): Promise<CourseWithProgress[]> {
    const res = await api.get<CourseWithProgress[]>(COURSES.list);
    return res.data;
  },

  async get(id: number): Promise<Course> {
    const res = await api.get<Course>(COURSES.byId(id));
    return res.data;
  },

  async listMaterials(courseId: number): Promise<TeachingMaterial[]> {
    const res = await api.get<TeachingMaterial[]>(COURSES.materials(courseId));
    return res.data;
  },

  /** コース内で最後に閲覧した章の閲覧記録を返す。履歴なし（204）のときは null。 */
  async lastViewed(courseId: number): Promise<UserChapterView | null> {
    const res = await api.get<UserChapterView | undefined>(COURSES.lastViewed(courseId));
    return res.status === 204 || !res.data ? null : res.data;
  },

  async create(payload: CoursePayload): Promise<Course> {
    const res = await api.post<Course>(COURSES.list, payload);
    return res.data;
  },

  async update(id: number, payload: CoursePayload): Promise<Course> {
    const res = await api.put<Course>(COURSES.byId(id), payload);
    return res.data;
  },

  async remove(id: number): Promise<void> {
    await api.delete(COURSES.byId(id));
  },
};

export default CourseRepository;
