import { useEffect, useState, useCallback } from 'react';
import { useSelector } from 'react-redux';
import { Navigate } from 'react-router-dom';
import AdminScenarioRepository, {
  AdminScenario,
  AdminScenarioForm,
} from '../repositories/AdminScenarioRepository';
import type { RootState } from '../store';
import Loading from '../components/Loading';
import PageIntro from '../components/ui/PageIntro';

const EMPTY_FORM: AdminScenarioForm = {
  name: '',
  description: '',
  category: '',
  roleName: '',
  difficulty: '',
  systemPrompt: '',
};

/**
 * 管理者専用: 練習シナリオ管理ページ。
 *
 * <p>Cognito の admin グループに所属しているユーザーのみアクセス可能。
 * 非 admin は /（ホーム）にリダイレクト。バックエンドも /api/admin/** で 403 を返すため二重防御。</p>
 */
export default function AdminScenariosPage() {
  const isAdmin = useSelector((state: RootState) => state.auth.isAdmin);
  const authLoading = useSelector((state: RootState) => state.auth.loading);

  const [scenarios, setScenarios] = useState<AdminScenario[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<number | null>(null);
  const [form, setForm] = useState<AdminScenarioForm>(EMPTY_FORM);
  const [submitting, setSubmitting] = useState(false);

  const fetchAll = useCallback(async () => {
    setLoading(true);
    try {
      const data = await AdminScenarioRepository.list();
      setScenarios(data);
      setError(null);
    } catch (e) {
      setError('シナリオの取得に失敗しました');
      // eslint-disable-next-line no-console
      console.error(e);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (isAdmin) {
      fetchAll();
    }
  }, [isAdmin, fetchAll]);

  // 認証ロード待ち
  if (authLoading) return <Loading message="認証情報を確認中..." />;
  // admin でなければホームへ
  if (!isAdmin) return <Navigate to="/" replace />;

  const startEdit = (s: AdminScenario) => {
    setEditingId(s.id);
    setForm({
      name: s.name,
      description: s.description ?? '',
      category: s.category ?? '',
      roleName: s.roleName,
      difficulty: s.difficulty ?? '',
      systemPrompt: s.systemPrompt,
    });
  };

  const reset = () => {
    setEditingId(null);
    setForm(EMPTY_FORM);
  };

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    try {
      if (editingId) {
        await AdminScenarioRepository.update(editingId, form);
      } else {
        await AdminScenarioRepository.create(form);
      }
      reset();
      await fetchAll();
    } catch (err) {
      setError('保存に失敗しました');
      // eslint-disable-next-line no-console
      console.error(err);
    } finally {
      setSubmitting(false);
    }
  };

  const remove = async (id: number) => {
    if (!confirm('このシナリオを削除しますか？')) return;
    try {
      await AdminScenarioRepository.remove(id);
      await fetchAll();
    } catch (err) {
      setError('削除に失敗しました');
      // eslint-disable-next-line no-console
      console.error(err);
    }
  };

  return (
    <div className="p-6 max-w-4xl mx-auto space-y-6">
      <PageIntro
        title="管理: 練習シナリオ"
        description={<>新規作成・編集・削除ができます。一般ユーザーには表示されません。</>}
      />

      {error && (
        <div role="alert" className="p-3 rounded border border-red-300 bg-red-50 text-red-800 text-sm">
          {error}
        </div>
      )}

      <form onSubmit={submit} className="space-y-3 p-4 border rounded-lg bg-[var(--color-surface-1)]">
        <h2 className="text-base font-bold">{editingId ? `シナリオ #${editingId} を編集` : '新規シナリオを作成'}</h2>

        <label className="block text-sm">
          <span className="block mb-1">名前 *</span>
          <input
            required
            maxLength={100}
            value={form.name}
            onChange={(e) => setForm({ ...form, name: e.target.value })}
            className="w-full border rounded px-2 py-1"
          />
        </label>

        <label className="block text-sm">
          <span className="block mb-1">役割名 *</span>
          <input
            required
            maxLength={100}
            value={form.roleName}
            onChange={(e) => setForm({ ...form, roleName: e.target.value })}
            className="w-full border rounded px-2 py-1"
          />
        </label>

        <div className="grid grid-cols-2 gap-3">
          <label className="block text-sm">
            <span className="block mb-1">カテゴリ</span>
            <input
              maxLength={50}
              value={form.category ?? ''}
              onChange={(e) => setForm({ ...form, category: e.target.value })}
              className="w-full border rounded px-2 py-1"
            />
          </label>
          <label className="block text-sm">
            <span className="block mb-1">難易度</span>
            <input
              maxLength={20}
              value={form.difficulty ?? ''}
              onChange={(e) => setForm({ ...form, difficulty: e.target.value })}
              className="w-full border rounded px-2 py-1"
            />
          </label>
        </div>

        <label className="block text-sm">
          <span className="block mb-1">説明</span>
          <textarea
            rows={2}
            value={form.description ?? ''}
            onChange={(e) => setForm({ ...form, description: e.target.value })}
            className="w-full border rounded px-2 py-1"
          />
        </label>

        <label className="block text-sm">
          <span className="block mb-1">System プロンプト *</span>
          <textarea
            required
            rows={6}
            value={form.systemPrompt}
            onChange={(e) => setForm({ ...form, systemPrompt: e.target.value })}
            className="w-full border rounded px-2 py-1 font-mono text-xs"
          />
        </label>

        <div className="flex gap-2">
          <button
            type="submit"
            disabled={submitting}
            className="px-4 py-2 rounded bg-emerald-600 text-white disabled:opacity-50"
          >
            {editingId ? '更新' : '作成'}
          </button>
          {editingId && (
            <button type="button" onClick={reset} className="px-4 py-2 rounded border">
              キャンセル
            </button>
          )}
        </div>
      </form>

      <section>
        <h2 className="text-base font-bold mb-3">登録済みシナリオ ({scenarios.length})</h2>
        {loading ? (
          <Loading message="読み込み中..." />
        ) : scenarios.length === 0 ? (
          <p className="text-sm text-[var(--color-text-muted)]">まだ登録されていません</p>
        ) : (
          <ul className="space-y-2">
            {scenarios.map((s) => (
              <li
                key={s.id}
                className="p-3 border rounded flex items-start justify-between gap-3 bg-[var(--color-surface-1)]"
              >
                <div className="flex-1 text-sm">
                  <p className="font-bold">
                    #{s.id} {s.name} <span className="text-xs text-[var(--color-text-muted)]">({s.roleName})</span>
                  </p>
                  <p className="text-xs text-[var(--color-text-muted)]">
                    {s.category ?? '-'} / {s.difficulty ?? '-'}
                  </p>
                  {s.description && <p className="text-xs mt-1">{s.description}</p>}
                </div>
                <div className="flex flex-col gap-1">
                  <button onClick={() => startEdit(s)} className="text-xs px-2 py-1 border rounded">
                    編集
                  </button>
                  <button onClick={() => remove(s.id)} className="text-xs px-2 py-1 border border-red-300 rounded text-red-700">
                    削除
                  </button>
                </div>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  );
}
