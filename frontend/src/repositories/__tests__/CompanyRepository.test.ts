import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('@/shared/api/axios', () => ({
  default: {
    get: vi.fn(),
    patch: vi.fn(),
  },
}));

import apiClient from '@/shared/api/axios';
import CompanyRepository from '../CompanyRepository';

const mockedGet = vi.mocked(apiClient.get);
const mockedPatch = vi.mocked(apiClient.patch);

describe('CompanyRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('list: 会社一覧を取得する', async () => {
    mockedGet.mockResolvedValue({
      data: [{ id: 1, name: 'A社', isActive: true, createdAt: '', updatedAt: '' }],
    });
    const companies = await CompanyRepository.list();
    expect(mockedGet).toHaveBeenCalledWith('/api/v2/admin/companies');
    expect(companies[0].isActive).toBe(true);
  });

  it('updateActive: 会社を無効化する PATCH を送る', async () => {
    mockedPatch.mockResolvedValue({ status: 200 });
    await CompanyRepository.updateActive(5, false);
    expect(mockedPatch).toHaveBeenCalledWith('/api/v2/admin/companies/5/active', { active: false });
  });
});
