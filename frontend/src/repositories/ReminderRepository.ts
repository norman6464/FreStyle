import apiClient from '../lib/axios';
import { ReminderSetting } from '../types';

export const ReminderRepository = {
  async fetchSetting(): Promise<ReminderSetting> {
    const response = await apiClient.get<ReminderSetting>('/api/v2/reminder');
    return response.data;
  },

  async saveSetting(setting: ReminderSetting): Promise<ReminderSetting> {
    const response = await apiClient.put<ReminderSetting>('/api/v2/reminder', setting);
    return response.data;
  },
};
