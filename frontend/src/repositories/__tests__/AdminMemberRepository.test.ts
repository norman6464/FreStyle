import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('../../lib/axios', () => ({
  default: {
    get: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
  },
}));

import apiClient from '../../lib/axios';
import AdminMemberRepository from '../AdminMemberRepository';

const mockedGet = vi.mocked(apiClient.get);
const mockedPatch = vi.mocked(apiClient.patch);
const mockedDelete = vi.mocked(apiClient.delete);

describe('AdminMemberRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('listMembers: 自社の従業員一覧を取得する', async () => {
    mockedGet.mockResolvedValue({
      data: [{ id: 1, email: 'a@e.com', displayName: '田中', role: 'trainee', aiChatEnabled: null }],
    });
    const members = await AdminMemberRepository.listMembers();
    expect(mockedGet).toHaveBeenCalledWith('/api/v2/admin/members');
    expect(members[0].displayName).toBe('田中');
  });

  it('updateAiAccess: 個別 OFF を送る', async () => {
    mockedPatch.mockResolvedValue({ status: 204 });
    await AdminMemberRepository.updateAiAccess(7, false);
    expect(mockedPatch).toHaveBeenCalledWith('/api/v2/admin/members/7/ai-access', { enabled: false });
  });

  it('updateAiAccess: null で会社設定に従う状態へ戻す', async () => {
    mockedPatch.mockResolvedValue({ status: 204 });
    await AdminMemberRepository.updateAiAccess(7, null);
    expect(mockedPatch).toHaveBeenCalledWith('/api/v2/admin/members/7/ai-access', { enabled: null });
  });

  it('updateActive: 従業員を無効化する PATCH を送る', async () => {
    mockedPatch.mockResolvedValue({ status: 204 });
    await AdminMemberRepository.updateActive(7, false);
    expect(mockedPatch).toHaveBeenCalledWith('/api/v2/admin/members/7/active', { active: false });
  });

  it('remove: 従業員を論理削除する DELETE を送る', async () => {
    mockedDelete.mockResolvedValue({ status: 204 });
    await AdminMemberRepository.remove(7);
    expect(mockedDelete).toHaveBeenCalledWith('/api/v2/admin/members/7');
  });
});
