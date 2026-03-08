import { describe, it, expect, vi, beforeEach } from 'vitest';
import apiClient from '../../lib/axios';
import { TemplateRepository } from '../TemplateRepository';

vi.mock('../../lib/axios');
const mockedApiClient = vi.mocked(apiClient);

describe('TemplateRepository', () => {
  beforeEach(() => vi.clearAllMocks());

  describe('fetchTemplates', () => {
    it('カテゴリなしで全テンプレートを取得する', async () => {
      mockedApiClient.get.mockResolvedValue({ data: [] });
      await TemplateRepository.fetchTemplates();
      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/templates', { params: {} });
    });

    it('カテゴリ指定でテンプレートを取得する', async () => {
      mockedApiClient.get.mockResolvedValue({ data: [] });
      await TemplateRepository.fetchTemplates('meeting');
      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/templates', { params: { category: 'meeting' } });
    });
  });

  describe('fetchTemplateById', () => {
    it('IDでテンプレートを取得する', async () => {
      const mockTemplate = { id: 1, title: 'Test', description: '', category: 'meeting', openingMessage: 'Hello', difficulty: 'beginner' };
      mockedApiClient.get.mockResolvedValue({ data: mockTemplate });
      const result = await TemplateRepository.fetchTemplateById(1);
      expect(mockedApiClient.get).toHaveBeenCalledWith('/api/templates/1');
      expect(result).toEqual(mockTemplate);
    });
  });
});
