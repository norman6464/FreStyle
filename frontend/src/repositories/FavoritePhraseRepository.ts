import type { FavoritePhrase } from '../types';

const STORAGE_KEY = 'freestyle_favorite_phrases';

export const FavoritePhraseRepository = {
  getAll(): FavoritePhrase[] {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    try {
      const phrases: FavoritePhrase[] = JSON.parse(raw);
      return phrases.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
    } catch {
      localStorage.removeItem(STORAGE_KEY);
      return [];
    }
  },

  save(phrase: Omit<FavoritePhrase, 'id' | 'createdAt'>): void {
    const all = this.getAll();
    const newPhrase: FavoritePhrase = {
      ...phrase,
      id: `fav_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`,
      createdAt: new Date(Date.now()).toISOString(),
    };
    all.push(newPhrase);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(all));
  },

  remove(id: string): void {
    const all = this.getAll();
    const filtered = all.filter((p) => p.id !== id);
    localStorage.setItem(STORAGE_KEY, JSON.stringify(filtered));
  },

  exists(rephrasedText: string, pattern: string): boolean {
    const all = this.getAll();
    return all.some((p) => p.rephrasedText === rephrasedText && p.pattern === pattern);
  },
};
