import apiClient from '../lib/axios';
import { ConversationTemplate } from '../types';

export const TemplateRepository = {
  async fetchTemplates(category?: string): Promise<ConversationTemplate[]> {
    const response = await apiClient.get<ConversationTemplate[]>('/api/templates', {
      params: category ? { category } : {},
    });
    return response.data;
  },

  async fetchTemplateById(id: number): Promise<ConversationTemplate> {
    const response = await apiClient.get<ConversationTemplate>(`/api/templates/${id}`);
    return response.data;
  },
};
