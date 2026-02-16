import apiClient from '../lib/axios';

const STORAGE_KEY = 'freestyle_scenario_bookmarks';

function getLocalBookmarks(): number[] {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return [];
  try {
    return JSON.parse(raw);
  } catch {
    localStorage.removeItem(STORAGE_KEY);
    return [];
  }
}

function saveLocalBookmarks(ids: number[]): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(ids));
}

export const BookmarkRepository = {
  async getAll(): Promise<number[]> {
    try {
      const response = await apiClient.get<number[]>('/api/bookmarks');
      return response.data;
    } catch {
      return getLocalBookmarks();
    }
  },

  async add(scenarioId: number): Promise<void> {
    try {
      await apiClient.post(`/api/bookmarks/${scenarioId}`);
    } catch {
      const ids = getLocalBookmarks();
      if (!ids.includes(scenarioId)) {
        ids.push(scenarioId);
        saveLocalBookmarks(ids);
      }
    }
  },

  async remove(scenarioId: number): Promise<void> {
    try {
      await apiClient.delete(`/api/bookmarks/${scenarioId}`);
    } catch {
      const ids = getLocalBookmarks().filter(id => id !== scenarioId);
      saveLocalBookmarks(ids);
    }
  },

  async isBookmarked(scenarioId: number): Promise<boolean> {
    const ids = await this.getAll();
    return ids.includes(scenarioId);
  },
};
