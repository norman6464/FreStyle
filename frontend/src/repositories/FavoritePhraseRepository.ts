import apiClient from '../lib/axios';
import type { FavoritePhrase } from '../types';

const STORAGE_KEY = 'freestyle_favorite_phrases';

function getLocalPhrases(): FavoritePhrase[] {
  const raw = localStorage.getItem(STORAGE_KEY);
  if (!raw) return [];
  try {
    const phrases: FavoritePhrase[] = JSON.parse(raw);
    return phrases.sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime());
  } catch {
    localStorage.removeItem(STORAGE_KEY);
    return [];
  }
}

function saveLocalPhrases(phrases: FavoritePhrase[]): void {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(phrases));
}

interface ApiFavoritePhraseDto {
  id: number;
  originalText: string;
  rephrasedText: string;
  pattern: string;
  createdAt: string;
}

function toFavoritePhrase(dto: ApiFavoritePhraseDto): FavoritePhrase {
  return {
    id: String(dto.id),
    originalText: dto.originalText,
    rephrasedText: dto.rephrasedText,
    pattern: dto.pattern,
    createdAt: dto.createdAt,
  };
}

export const FavoritePhraseRepository = {
  async getAll(): Promise<FavoritePhrase[]> {
    try {
      const response = await apiClient.get<ApiFavoritePhraseDto[]>('/api/favorite-phrases');
      return response.data.map(toFavoritePhrase);
    } catch {
      return getLocalPhrases();
    }
  },

  async save(phrase: Omit<FavoritePhrase, 'id' | 'createdAt'>): Promise<void> {
    try {
      await apiClient.post('/api/favorite-phrases', {
        originalText: phrase.originalText,
        rephrasedText: phrase.rephrasedText,
        pattern: phrase.pattern,
      });
    } catch {
      const all = getLocalPhrases();
      const newPhrase: FavoritePhrase = {
        ...phrase,
        id: `fav_${Date.now()}_${Math.random().toString(36).slice(2, 8)}`,
        createdAt: new Date(Date.now()).toISOString(),
      };
      all.push(newPhrase);
      saveLocalPhrases(all);
    }
  },

  async remove(id: string): Promise<void> {
    try {
      await apiClient.delete(`/api/favorite-phrases/${id}`);
    } catch {
      const all = getLocalPhrases().filter((p) => p.id !== id);
      saveLocalPhrases(all);
    }
  },

  async exists(rephrasedText: string, pattern: string): Promise<boolean> {
    const all = await this.getAll();
    return all.some((p) => p.rephrasedText === rephrasedText && p.pattern === pattern);
  },
};
