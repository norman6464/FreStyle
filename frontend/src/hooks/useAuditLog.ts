import { useEffect, useState } from 'react';
import { AuditRepository, AuditEvent } from '@/entities/audit';

/**
 * useAuditLog — 監査ログ（super_admin）を取得するフック。
 * `enabled=false`（super_admin 以外 / 認証確認中）のときは API を叩かない。
 */
export function useAuditLog(enabled = true) {
  const [events, setEvents] = useState<AuditEvent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!enabled) {
      setLoading(false);
      return;
    }
    let cancelled = false;
    // 取得開始時に loading/error を初期化（enabled が false→true に変わる再取得でも整合を取る）。
    setLoading(true);
    setError(null);
    AuditRepository.list()
      .then((e) => {
        if (!cancelled) setEvents(e);
      })
      .catch(() => {
        if (!cancelled) setError('監査ログの取得に失敗しました');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, [enabled]);

  return { events, loading, error };
}
