import api from '../lib/axios';
import { COURSES } from '../constants/apiRoutes';
import type { Course, TeachingMaterial } from '../types';

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
  sortOrder: number;
  isPublished: boolean;
}

const CourseRepository = {
  async list(): Promise<Course[]> {
    const res = await api.get<Course[]>(COURSES.list);
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
