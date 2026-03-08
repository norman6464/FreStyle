import { useState, useEffect, useCallback } from 'react';
import { SharedSessionRepository } from '../repositories/SharedSessionRepository';
import { SharedSession } from '../types';

export function useSharedSessions() {
  const [sessions, setSessions] = useState<SharedSession[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    SharedSessionRepository.fetchPublicSessions()
      .then((data) => {
        if (!cancelled) setSessions(data);
      })
      .catch(() => {
        if (!cancelled) setError('共有セッションの取得に失敗しました');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => { cancelled = true; };
  }, []);

  const shareSession = useCallback(async (sessionId: number, description?: string) => {
    const shared = await SharedSessionRepository.shareSession(sessionId, description);
    setSessions((prev) => [shared, ...prev]);
    return shared;
  }, []);

  const unshareSession = useCallback(async (sessionId: number) => {
    await SharedSessionRepository.unshareSession(sessionId);
    setSessions((prev) => prev.filter((s) => s.sessionId !== sessionId));
  }, []);

  return { sessions, loading, error, shareSession, unshareSession };
}
