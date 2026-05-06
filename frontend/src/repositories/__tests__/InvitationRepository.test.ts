import { describe, it, expect, vi, beforeEach } from 'vitest';
import invitationRepository from '../InvitationRepository';
import apiClient from '../../lib/axios';

vi.mock('../../lib/axios');

const mockedApiClient = vi.mocked(apiClient);

describe('InvitationRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('validateToken: token を URL に埋め込んで GET する', async () => {
    const mockInv = {
      role: 'company_admin',
      displayName: '山田',
      companyId: 42,
      companyName: '株式会社FreStyle',
    };
    mockedApiClient.get.mockResolvedValue({ data: mockInv });

    const result = await invitationRepository.validateToken('abc-123');

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/v2/invitations/accept/abc-123');
    expect(result).toEqual(mockInv);
  });

  it('validateToken: メタ文字を含む token は URL エンコードされる', async () => {
    mockedApiClient.get.mockResolvedValue({ data: {} });
    await invitationRepository.validateToken('a/b?x=1');
    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/v2/invitations/accept/a%2Fb%3Fx%3D1');
  });
});
