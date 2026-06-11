import { useCallback, useState } from 'react';
import axios from 'axios';
import AdminSqlRepository, { SqlResult } from '../repositories/AdminSqlRepository';

/**
 * useAdminSql — super_admin 向け read-only SQL コンソールの状態管理フック。
 * クエリ実行と、結果 / 実行中 / エラー（SQL エラーや読み取り専用違反は backend のメッセージをそのまま表示）を扱う。
 */
export function useAdminSql() {
  const [result, setResult] = useState<SqlResult | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const run = useCallback(async (query: string) => {
    setLoading(true);
    setError(null);
    try {
      setResult(await AdminSqlRepository.run(query));
    } catch (e) {
      setResult(null);
      const msg = axios.isAxiosError(e)
        ? (e.response?.data as { error?: string } | undefined)?.error
        : undefined;
      setError(msg || 'クエリの実行に失敗しました');
    } finally {
      setLoading(false);
    }
  }, []);

  return { result, loading, error, run };
}
