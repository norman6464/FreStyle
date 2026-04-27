import { describe, it, expect, vi, beforeEach } from 'vitest';
import apiClient from '../../lib/axios';
import { ReminderRepository } from '../ReminderRepository';

vi.mock('../../lib/axios');
const mockedApiClient = vi.mocked(apiClient);

describe('ReminderRepository', () => {
  beforeEach(() => vi.clearAllMocks());

  it('設定を取得する', async () => {
    const mockData = { enabled: true, reminderTime: '20:00', daysOfWeek: 'mon,tue' };
    mockedApiClient.get.mockResolvedValue({ data: mockData });
    const result = await ReminderRepository.fetchSetting();
    expect(mockedApiClient.get).toHaveBeenCalledWith('/api/v2/reminder');
    expect(result).toEqual(mockData);
  });

  it('設定を保存する', async () => {
    const setting = { enabled: true, reminderTime: '19:00', daysOfWeek: 'mon' };
    mockedApiClient.put.mockResolvedValue({ data: setting });
    const result = await ReminderRepository.saveSetting(setting);
    expect(mockedApiClient.put).toHaveBeenCalledWith('/api/v2/reminder', setting);
    expect(result).toEqual(setting);
  });
});
