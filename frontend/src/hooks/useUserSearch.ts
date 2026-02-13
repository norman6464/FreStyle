import { useState, useEffect, useMemo } from 'react';
import { debounce } from 'lodash';
import UserSearchRepository from '../repositories/UserSearchRepository';
import type { MemberUser } from '../types';

export function useUserSearch() {
  const [users, setUsers] = useState<MemberUser[]>([]);
  const [error, setError] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const [debounceQuery, setDebounceQuery] = useState('');

  const debounceSearch = useMemo(
    () => debounce((query: string) => setDebounceQuery(query), 500),
    []
  );

  useEffect(() => {
    debounceSearch(searchQuery);
    return () => debounceSearch.cancel();
  }, [searchQuery, debounceSearch]);

  useEffect(() => {
    let cancelled = false;

    const search = async () => {
      try {
        const result = await UserSearchRepository.searchUsers(debounceQuery || undefined);
        if (!cancelled) {
          setUsers(result);
          setError(null);
        }
      } catch (err) {
        if (!cancelled) {
          setError((err as Error).message);
        }
      }
    };

    search();

    return () => {
      cancelled = true;
    };
  }, [debounceQuery]);

  return {
    users,
    error,
    searchQuery,
    setSearchQuery,
    debounceQuery,
  };
}
