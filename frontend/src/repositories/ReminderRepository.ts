import apiClient from '../lib/axios';
import { REMINDER } from '../constants/apiRoutes';
import { ReminderSetting } from '../types';

export const ReminderRepository = {
  async fetchSetting(): Promise<ReminderSetting> {
    const response = await apiClient.get<ReminderSetting>(REMINDER);
    return response.data;
  },

  async saveSetting(setting: ReminderSetting): Promise<ReminderSetting> {
    const response = await apiClient.put<ReminderSetting>(REMINDER, setting);
    return response.data;
  },
};
