import apiClient from '../lib/axios';
import { TEMPLATES } from '../constants/apiRoutes';
import { ConversationTemplate } from '../types';

export const TemplateRepository = {
  async fetchTemplates(category?: string): Promise<ConversationTemplate[]> {
    const response = await apiClient.get<ConversationTemplate[]>(TEMPLATES.list, {
      params: category ? { category } : {},
    });
    return response.data;
  },

  async fetchTemplateById(id: number): Promise<ConversationTemplate> {
    const response = await apiClient.get<ConversationTemplate>(TEMPLATES.byId(id));
    return response.data;
  },
};
