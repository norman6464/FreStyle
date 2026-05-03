import apiClient from '../lib/axios';
import { ADMIN } from '../constants/apiRoutes';

export interface Company {
  id: number;
  name: string;
  createdAt: string;
  updatedAt: string;
}

class CompanyRepository {
  async list(): Promise<Company[]> {
    const res = await apiClient.get<Company[]>(ADMIN.companies);
    return res.data;
  }
}

export default new CompanyRepository();
