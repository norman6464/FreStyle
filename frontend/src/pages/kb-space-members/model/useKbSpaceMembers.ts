import { useCallback, useEffect, useState } from 'react';
import { KbRepository, type KbSpaceMember } from '@/entities/kb';

export function useKbSpaceMembers(workspaceSlug: string, spaceId: string) {
  const [members, setMembers] = useState<KbSpaceMember[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(() => {
    setLoading(true);
    setError(null);
    KbRepository.fetchSpaceMembers(workspaceSlug, spaceId)
      .then(setMembers)
      .catch(() => setError('メンバーを読み込めませんでした。'))
      .finally(() => setLoading(false));
  }, [workspaceSlug, spaceId]);

  useEffect(() => {
    load();
  }, [load]);

  return { members, loading, error, retry: load };
}
