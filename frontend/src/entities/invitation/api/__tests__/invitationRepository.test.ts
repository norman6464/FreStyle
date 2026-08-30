import { describe, it, expect, vi, beforeEach } from 'vitest';
import invitationRepository from '../invitationRepository';
import apiClient from '@/shared/api/axios';

vi.mock('@/shared/api/axios');

const mockedApiClient = vi.mocked(apiClient);

describe('InvitationRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('validateToken: token を URL に埋め込んで GET する', async () => {
    // backend が返すのは表示名のキーが name の形（companyId は返さない）。
    const mockInv = {
      role: 'company_admin',
      name: '山田',
      companyName: '株式会社FreStyle',
    };
    mockedApiClient.get.mockResolvedValue({ data: mockInv });

    const result = await invitationRepository.validateToken('abc-123');

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/v2/invitations/accept/abc-123');
    // name を displayName へ読み替える。素通しのままだと受諾画面の表示名が空になる。
    expect(result).toEqual({
      role: 'company_admin',
      displayName: '山田',
      companyName: '株式会社FreStyle',
    });
  });

  it('validateToken: メタ文字を含む token は URL エンコードされる', async () => {
    mockedApiClient.get.mockResolvedValue({ data: {} });
    await invitationRepository.validateToken('a/b?x=1');
    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/v2/invitations/accept/a%2Fb%3Fx%3D1');
  });
});
