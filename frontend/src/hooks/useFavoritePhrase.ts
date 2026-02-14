import { useState, useCallback, useMemo } from 'react';
import { FavoritePhraseRepository } from '../repositories/FavoritePhraseRepository';
import type { FavoritePhrase } from '../types';

export function useFavoritePhrase() {
  const [phrases, setPhrases] = useState<FavoritePhrase[]>(() => FavoritePhraseRepository.getAll());
  const [searchQuery, setSearchQuery] = useState('');
  const [patternFilter, setPatternFilter] = useState('すべて');

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

  const filteredPhrases = useMemo(() => {
    return phrases.filter((phrase) => {
      const matchesPattern = patternFilter === 'すべて' || phrase.pattern === patternFilter;
      const matchesSearch = !searchQuery ||
        phrase.originalText.includes(searchQuery) ||
        phrase.rephrasedText.includes(searchQuery);
      return matchesPattern && matchesSearch;
    });
  }, [phrases, searchQuery, patternFilter]);

  return { phrases, filteredPhrases, searchQuery, setSearchQuery, patternFilter, setPatternFilter, saveFavorite, removeFavorite, isFavorite };
}
