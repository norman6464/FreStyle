import { useCompanyAiSettings } from '../../hooks/useCompanyAiSettings';

/**
 * 会社設定: trainee への AI エージェント機能の有効/無効トグル。
 * company_admin / super_admin のみがアクセスする設定セクションで描画される。
 */
export default function CompanyAiSettings() {
  const { enabled, loading, saving, error, update } = useCompanyAiSettings();

  return (
    <div className="max-w-2xl mx-auto p-6">
      <div className="rounded-xl border border-surface-3 bg-surface-1 p-6">
        <h2 className="text-lg font-bold text-[var(--color-text-primary)]">AI エージェント設定</h2>
        <p className="mt-1 text-sm text-[var(--color-text-muted)]">
          受講者（trainee）が AI エージェント機能（AI チャット）を利用できるかどうかを切り替えます。
        </p>

        {error && (
          <div className="mt-4 rounded-md border border-red-500/30 bg-red-500/5 p-3 text-sm text-red-400">
            {error}
          </div>
        )}

        <label className="mt-6 flex items-start gap-3 cursor-pointer select-none">
          <input
            type="checkbox"
            className="mt-1 h-4 w-4 accent-[var(--color-accent)] cursor-pointer disabled:cursor-not-allowed"
            checked={!!enabled}
            disabled={loading || saving}
            onChange={(e) => update(e.target.checked)}
            aria-label="受講者の AI エージェント機能を有効にする"
          />
          <span>
            <span className="block text-sm font-medium text-[var(--color-text-primary)]">
              受講者の AI エージェント機能を有効にする
            </span>
            <span className="block text-xs text-[var(--color-text-muted)] mt-0.5">
              無効にすると、受講者のサイドバーから「AI」が消え、AI チャットの利用ができなくなります。
              （管理者の利用には影響しません）
            </span>
          </span>
        </label>

        {(loading || saving) && (
          <p className="mt-3 text-xs text-[var(--color-text-muted)]">
            {loading ? '読み込み中…' : '保存中…'}
          </p>
        )}
      </div>
    </div>
  );
}
