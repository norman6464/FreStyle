import { useState, useCallback } from 'react';
import { BookmarkRepository } from '../repositories/BookmarkRepository';

export function useBookmark() {
  const [bookmarkedIds, setBookmarkedIds] = useState<number[]>(() => BookmarkRepository.getAll());

  const toggleBookmark = useCallback((scenarioId: number) => {
    if (BookmarkRepository.isBookmarked(scenarioId)) {
      BookmarkRepository.remove(scenarioId);
    } else {
      BookmarkRepository.add(scenarioId);
    }
    setBookmarkedIds(BookmarkRepository.getAll());
  }, []);

  const isBookmarked = useCallback((scenarioId: number) => {
    return BookmarkRepository.isBookmarked(scenarioId);
  }, []);

  return { bookmarkedIds, toggleBookmark, isBookmarked };
}
