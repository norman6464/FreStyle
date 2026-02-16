import { useState, useCallback, useMemo, useEffect } from 'react';
import { FavoritePhraseRepository } from '../repositories/FavoritePhraseRepository';
import type { FavoritePhrase } from '../types';

export function useFavoritePhrase() {
  const [phrases, setPhrases] = useState<FavoritePhrase[]>([]);
  const [loading, setLoading] = useState(true);
  const [searchQuery, setSearchQuery] = useState('');
  const [patternFilter, setPatternFilter] = useState('すべて');

  useEffect(() => {
    let cancelled = false;
    FavoritePhraseRepository.getAll().then((data) => {
      if (!cancelled) {
        setPhrases(data);
        setLoading(false);
      }
    });
    return () => { cancelled = true; };
  }, []);

  const saveFavorite = useCallback(async (originalText: string, rephrasedText: string, pattern: string) => {
    await FavoritePhraseRepository.save({ originalText, rephrasedText, pattern });
    const updated = await FavoritePhraseRepository.getAll();
    setPhrases(updated);
  }, []);

  const removeFavorite = useCallback(async (id: string) => {
    setPhrases(prev => prev.filter(p => p.id !== id));
    await FavoritePhraseRepository.remove(id);
  }, []);

  const isFavorite = useCallback((rephrasedText: string, pattern: string) => {
    return phrases.some((p) => p.rephrasedText === rephrasedText && p.pattern === pattern);
  }, [phrases]);

  const filteredPhrases = useMemo(() => {
    return phrases.filter((phrase) => {
      const matchesPattern = patternFilter === 'すべて' || phrase.pattern === patternFilter;
      const matchesSearch = !searchQuery ||
        phrase.originalText.includes(searchQuery) ||
        phrase.rephrasedText.includes(searchQuery);
      return matchesPattern && matchesSearch;
    });
  }, [phrases, searchQuery, patternFilter]);

  return { phrases, filteredPhrases, searchQuery, setSearchQuery, patternFilter, setPatternFilter, saveFavorite, removeFavorite, isFavorite, loading };
}
