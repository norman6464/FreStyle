import { useCallback, useEffect, useState } from 'react';
import { KbRepository, type KbFavoritePage } from '@/entities/kb';

export function useKbFavorites(workspaceSlug: string) {
  const [favorites, setFavorites] = useState<KbFavoritePage[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const load = useCallback(() => {
    setLoading(true);
    setError(null);
    KbRepository.fetchFavorites(workspaceSlug)
      .then(setFavorites)
      .catch(() => setError('お気に入りを読み込めませんでした。'))
      .finally(() => setLoading(false));
  }, [workspaceSlug]);

  useEffect(() => {
    load();
  }, [load]);

  return { favorites, loading, error, retry: load };
}
