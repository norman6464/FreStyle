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
    const snapshot = bookmarkedIds;
    if (isCurrentlyBookmarked) {
      setBookmarkedIds(prev => prev.filter(id => id !== scenarioId));
      try {
        await BookmarkRepository.remove(scenarioId);
      } catch {
        setBookmarkedIds(snapshot);
      }
    } else {
      setBookmarkedIds(prev => [...prev, scenarioId]);
      try {
        await BookmarkRepository.add(scenarioId);
      } catch {
        setBookmarkedIds(snapshot);
      }
    }
  }, [bookmarkedIds]);

  const isBookmarked = useCallback((scenarioId: number) => {
    return bookmarkedIds.includes(scenarioId);
  }, [bookmarkedIds]);

  return { bookmarkedIds, toggleBookmark, isBookmarked, loading };
}
