import { useState, useEffect } from 'react';
import { BookmarkRepository } from '../repositories/BookmarkRepository';
import PracticeRepository, { type PracticeScenario } from '../repositories/PracticeRepository';

export function useBookmarkedScenarios(limit = 3) {
  const [scenarios, setScenarios] = useState<PracticeScenario[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setScenarios([]);

    Promise.all([BookmarkRepository.getAll(), PracticeRepository.getScenarios()])
      .then(([bookmarkedIds, allScenarios]) => {
        if (cancelled) return;
        const bookmarkedIdSet = new Set(bookmarkedIds);
        const bookmarked = allScenarios.filter((s) => bookmarkedIdSet.has(s.id));
        setScenarios(bookmarked.slice(0, limit));
        setLoading(false);
      })
      .catch(() => {
        if (!cancelled) setLoading(false);
      });

    return () => { cancelled = true; };
  }, [limit]);

  return { scenarios, loading };
}
