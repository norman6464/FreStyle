import api from '../lib/axios';
import { TEACHING_MATERIALS } from '../constants/apiRoutes';
import type { TeachingMaterial } from '../types';

/**
 * 教材 API ラッパ。
 *
 * actor の role / company は backend 側で context から取り出して自動フィルタするため、
 * フロントは company_id を渡さない（IDOR 対策）。
 */
export interface TeachingMaterialPayload {
  title: string;
  content: string;
  isPublished: boolean;
}

const TeachingMaterialRepository = {
  async list(): Promise<TeachingMaterial[]> {
    const res = await api.get<TeachingMaterial[]>(TEACHING_MATERIALS.list);
    return res.data;
  },

  async get(id: number): Promise<TeachingMaterial> {
    const res = await api.get<TeachingMaterial>(TEACHING_MATERIALS.byId(id));
    return res.data;
  },

  async create(payload: TeachingMaterialPayload): Promise<TeachingMaterial> {
    const res = await api.post<TeachingMaterial>(TEACHING_MATERIALS.list, payload);
    return res.data;
  },

  async update(id: number, payload: TeachingMaterialPayload): Promise<TeachingMaterial> {
    const res = await api.put<TeachingMaterial>(TEACHING_MATERIALS.byId(id), payload);
    return res.data;
  },

  async remove(id: number): Promise<void> {
    await api.delete(TEACHING_MATERIALS.byId(id));
  },
};

export default TeachingMaterialRepository;
