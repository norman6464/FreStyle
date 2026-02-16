const STORAGE_KEY = 'freestyle_scenario_bookmarks';

export const BookmarkRepository = {
  getAll(): number[] {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    try {
      return JSON.parse(raw);
    } catch {
      localStorage.removeItem(STORAGE_KEY);
      return [];
    }
  },

  add(scenarioId: number): void {
    const ids = this.getAll();
    if (!ids.includes(scenarioId)) {
      ids.push(scenarioId);
      localStorage.setItem(STORAGE_KEY, JSON.stringify(ids));
    }
  },

  remove(scenarioId: number): void {
    const ids = this.getAll().filter(id => id !== scenarioId);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(ids));
  },

  isBookmarked(scenarioId: number): boolean {
    return this.getAll().includes(scenarioId);
  },
};
