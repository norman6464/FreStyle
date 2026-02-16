import { useState, useEffect } from 'react';
import { BookmarkRepository } from '../repositories/BookmarkRepository';
import PracticeRepository, { type PracticeScenario } from '../repositories/PracticeRepository';

export function useBookmarkedScenarios(limit = 3) {
  const [scenarios, setScenarios] = useState<PracticeScenario[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;

    Promise.all([BookmarkRepository.getAll(), PracticeRepository.getScenarios()])
      .then(([bookmarkedIds, allScenarios]) => {
        if (cancelled) return;
        const bookmarked = allScenarios.filter((s) => bookmarkedIds.includes(s.id));
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
