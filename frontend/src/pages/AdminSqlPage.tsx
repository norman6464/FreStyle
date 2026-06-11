import { useState, useCallback, KeyboardEvent } from 'react';
import { useSelector } from 'react-redux';
import { Navigate } from 'react-router-dom';
import type { RootState } from '../store';
import Loading from '../components/Loading';
import PageIntro from '../components/ui/PageIntro';
import { useAdminSql } from '../hooks/useAdminSql';

// セルの表示。null は NULL、boolean は true/false を明示する。
function renderCell(v: string | number | boolean | null): string {
  if (v === null) return 'NULL';
  if (typeof v === 'boolean') return v ? 'true' : 'false';
  return String(v);
}

/**
 * AdminSqlPage — `/admin/sql`。super_admin（運営管理者）専用の read-only SQL コンソール。
 * 本番 DB に対して SELECT / WITH クエリのみを実行できる（書き込みは backend / DB 側で拒否）。
 */
export default function AdminSqlPage() {
  const role = useSelector((s: RootState) => s.auth.role);
  const authLoading = useSelector((s: RootState) => s.auth.loading);
  const { result, loading, error, run } = useAdminSql();
  const [query, setQuery] = useState('SELECT id, email, role FROM users ORDER BY id LIMIT 20;');

  const onRun = useCallback(() => {
    const q = query.trim();
    if (q && !loading) run(q);
  }, [query, loading, run]);

  const onKeyDown = useCallback(
    (e: KeyboardEvent<HTMLTextAreaElement>) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
        e.preventDefault();
        onRun();
      }
    },
    [onRun],
  );

  if (authLoading) return <Loading message="読み込み中..." className="min-h-[50vh]" />;
  // 運営管理者(super_admin)専用。company_admin / trainee は弾く。
  if (role !== 'super_admin') return <Navigate to="/" replace />;

  return (
    <div className="px-4 sm:px-6 pt-6 pb-24 max-w-6xl mx-auto">
      <PageIntro
        title="運営: SQL コンソール（読み取り専用）"
        description="本番 DB に対して SELECT / WITH クエリだけを実行できます。書き込み（INSERT/UPDATE/DELETE 等）は実行できません。結果は最大 1000 行で打ち切られます。"
      />

      <div className="mt-4 rounded-md border border-amber-300 bg-amber-50 px-3 py-2 text-sm text-amber-800">
        ⚠️ 運営管理者専用の読み取り専用ツールです。実行内容（誰がどのクエリを実行したか）は監査ログに記録されます。
      </div>

      <textarea
        aria-label="SQL クエリ"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        onKeyDown={onKeyDown}
        spellCheck={false}
        rows={6}
        placeholder="SELECT ..."
        className="mt-4 w-full font-mono text-sm rounded-lg bg-surface-2 border border-surface-3 px-3 py-2 text-[var(--color-text-primary)] focus:outline-none focus:border-brand-500"
      />

      <div className="mt-2 flex items-center gap-3">
        <button
          type="button"
          onClick={onRun}
          disabled={loading || query.trim() === ''}
          className="px-5 py-2 rounded-lg bg-brand-500 text-white text-sm font-medium hover:bg-brand-600 active:bg-brand-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? '実行中...' : '実行'}
        </button>
        <span className="text-xs text-[var(--color-text-muted)]">⌘ / Ctrl + Enter で実行</span>
      </div>

      {error && (
        <pre className="mt-4 overflow-x-auto whitespace-pre-wrap rounded-md border border-rose-300 bg-rose-50 px-3 py-2 text-sm text-rose-700">
          {error}
        </pre>
      )}

      {result && !error && (
        <div className="mt-6">
          <div className="mb-2 text-sm text-[var(--color-text-muted)]">
            {result.rowCount} 行
            {result.truncated && '（上限の 1000 行で打ち切られました）'}
          </div>
          {result.columns.length === 0 ? (
            <p className="text-[var(--color-text-muted)]">結果なし（0 列）</p>
          ) : (
            <div className="overflow-x-auto rounded-lg border border-surface-3">
              <table className="w-full text-sm">
                <thead>
                  <tr className="bg-surface-2 text-left text-[var(--color-text-muted)]">
                    {result.columns.map((col) => (
                      <th key={col} className="px-3 py-2 font-medium whitespace-nowrap">
                        {col}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {result.rows.map((row, ri) => (
                    <tr key={ri} className="border-t border-surface-3 align-top">
                      {row.map((cell, ci) => (
                        <td
                          key={ci}
                          className="px-3 py-2 font-mono text-xs text-[var(--color-text-primary)] whitespace-pre-wrap break-all"
                        >
                          {renderCell(cell)}
                        </td>
                      ))}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
