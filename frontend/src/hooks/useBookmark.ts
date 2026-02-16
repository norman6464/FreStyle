import { useState, useCallback, useEffect } from 'react';
import { BookmarkRepository } from '../repositories/BookmarkRepository';

export function useBookmark() {
  const [bookmarkedIds, setBookmarkedIds] = useState<number[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    BookmarkRepository.getAll().then((ids) => {
      if (!cancelled) {
        setBookmarkedIds(ids);
        setLoading(false);
      }
    });
    return () => { cancelled = true; };
  }, []);

  const toggleBookmark = useCallback(async (scenarioId: number) => {
    const isCurrentlyBookmarked = bookmarkedIds.includes(scenarioId);
    if (isCurrentlyBookmarked) {
      setBookmarkedIds(prev => prev.filter(id => id !== scenarioId));
      await BookmarkRepository.remove(scenarioId);
    } else {
      setBookmarkedIds(prev => [...prev, scenarioId]);
      await BookmarkRepository.add(scenarioId);
    }
  }, [bookmarkedIds]);

  const isBookmarked = useCallback((scenarioId: number) => {
    return bookmarkedIds.includes(scenarioId);
  }, [bookmarkedIds]);

  return { bookmarkedIds, toggleBookmark, isBookmarked, loading };
}
