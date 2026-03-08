import { useState, useEffect, useCallback } from 'react';
import { WeeklyChallengeRepository } from '../repositories/WeeklyChallengeRepository';
import { WeeklyChallenge } from '../types';

export function useWeeklyChallenge() {
  const [challenge, setChallenge] = useState<WeeklyChallenge | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    let cancelled = false;
    WeeklyChallengeRepository.fetchCurrentChallenge()
      .then((data) => {
        if (!cancelled) setChallenge(data);
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => { cancelled = true; };
  }, []);

  const incrementProgress = useCallback(async () => {
    const updated = await WeeklyChallengeRepository.incrementProgress();
    if (updated) setChallenge(updated);
  }, []);

  return { challenge, loading, incrementProgress };
}
