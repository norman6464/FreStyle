import { useState, useCallback } from 'react';
import { FavoritePhraseRepository } from '../repositories/FavoritePhraseRepository';
import type { FavoritePhrase } from '../types';

export function useFavoritePhrase() {
  const [phrases, setPhrases] = useState<FavoritePhrase[]>(() => FavoritePhraseRepository.getAll());

  const saveFavorite = useCallback((originalText: string, rephrasedText: string, pattern: string) => {
    FavoritePhraseRepository.save({ originalText, rephrasedText, pattern });
    setPhrases(FavoritePhraseRepository.getAll());
  }, []);

  const removeFavorite = useCallback((id: string) => {
    FavoritePhraseRepository.remove(id);
    setPhrases(FavoritePhraseRepository.getAll());
  }, []);

  const isFavorite = useCallback((rephrasedText: string, pattern: string) => {
    return FavoritePhraseRepository.exists(rephrasedText, pattern);
  }, []);

  return { phrases, saveFavorite, removeFavorite, isFavorite };
}
