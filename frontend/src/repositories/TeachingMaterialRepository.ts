import api from '../lib/axios';
import { TEACHING_MATERIALS } from '../constants/apiRoutes';
import type { TeachingMaterial } from '../types';

/**
 * 教材 API ラッパ（個別 CRUD）。
 *
 * 一覧取得はコース配下なので CourseRepository.listMaterials を使う。
 * actor の role / company は backend 側で context から取り出して自動フィルタするため、
 * フロントは companyId を渡さない（IDOR 対策）。
 */
export interface TeachingMaterialCreatePayload {
  /** 所属コース ID（必須）。 */
  courseId: number;
  title: string;
  content: string;
  orderInCourse: number;
  isPublished: boolean;
}

export interface TeachingMaterialUpdatePayload {
  title: string;
  content: string;
  orderInCourse: number;
  isPublished: boolean;
}

const TeachingMaterialRepository = {
  async get(id: number): Promise<TeachingMaterial> {
    const res = await api.get<TeachingMaterial>(TEACHING_MATERIALS.byId(id));
    return res.data;
  },

  async create(payload: TeachingMaterialCreatePayload): Promise<TeachingMaterial> {
    const res = await api.post<TeachingMaterial>(TEACHING_MATERIALS.create, payload);
    return res.data;
  },

  async update(id: number, payload: TeachingMaterialUpdatePayload): Promise<TeachingMaterial> {
    const res = await api.put<TeachingMaterial>(TEACHING_MATERIALS.byId(id), payload);
    return res.data;
  },

  async remove(id: number): Promise<void> {
    await api.delete(TEACHING_MATERIALS.byId(id));
  },
};

export default TeachingMaterialRepository;
