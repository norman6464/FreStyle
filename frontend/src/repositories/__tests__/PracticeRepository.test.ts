import { describe, it, expect, vi, beforeEach } from 'vitest';
import practiceRepository from '../PracticeRepository';
import apiClient from '../../lib/axios';

vi.mock('../../lib/axios');

const mockedApiClient = vi.mocked(apiClient);

describe('PracticeRepository', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('getScenarios: シナリオ一覧を取得できる', async () => {
    const mockScenarios = [
      { id: 1, name: 'テストシナリオ', description: '説明', category: 'customer', roleName: '顧客', difficulty: 'easy', systemPrompt: 'prompt' },
    ];
    mockedApiClient.get.mockResolvedValue({ data: mockScenarios });

    const result = await practiceRepository.getScenarios();

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/practice/scenarios');
    expect(result).toEqual(mockScenarios);
  });

  it('getScenario: シナリオ詳細を取得できる', async () => {
    const mockScenario = { id: 1, name: 'テストシナリオ', description: '説明', category: 'customer', roleName: '顧客', difficulty: 'easy', systemPrompt: 'prompt' };
    mockedApiClient.get.mockResolvedValue({ data: mockScenario });

    const result = await practiceRepository.getScenario(1);

    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/practice/scenarios/1');
    expect(result).toEqual(mockScenario);
  });

  it('createPracticeSession: 練習セッションを作成できる', async () => {
    const mockSession = { id: 1, scenarioId: 1 };
    mockedApiClient.post.mockResolvedValue({ data: mockSession });

    const result = await practiceRepository.createPracticeSession({ scenarioId: 1 });

    expect(mockedApiClient.post).toHaveBeenCalledWith('/api/practice/sessions', { scenarioId: 1 });
    expect(result).toEqual(mockSession);
  });
});
