import type { LearningReport } from '../model/types';

interface ReportCardProps {
  report: LearningReport;
}

// 生成状態のラベルと配色。淡色背景 + 濃色文字でライトテーマでも読める。
const STATUS_STYLE: Record<LearningReport['status'], { label: string; className: string }> = {
  pending: { label: '作成中', className: 'text-amber-700 bg-amber-50 border-amber-200' },
  ready: { label: '完成', className: 'text-emerald-700 bg-emerald-50 border-emerald-200' },
  failed: { label: '失敗', className: 'text-rose-700 bg-rose-50 border-rose-200' },
};

/** 対象月・生成状態・作成日を 1 行で示す学習レポートのカード。 */
export default function ReportCard({ report }: ReportCardProps) {
  const status = STATUS_STYLE[report.status] ?? STATUS_STYLE.pending;

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4 flex items-center justify-between gap-3">
      <div className="min-w-0">
        <h3 className="text-sm font-semibold text-[var(--color-text-primary)]">
          {formatPeriod(report.periodFrom)} の学習レポート
        </h3>
        <p className="mt-0.5 text-xs text-[var(--color-text-muted)]">
          作成日: {formatDate(report.createdAt)}
        </p>
      </div>
      <span
        className={`shrink-0 text-xs font-medium px-2 py-0.5 rounded-full border ${status.className}`}
      >
        {status.label}
      </span>
    </div>
  );
}

/** ISO 文字列（対象期間の月初、UTC）から「YYYY年M月」を作る。不正値はそのまま返す。
 *  期間境界は UTC 深夜のため、UTC メソッドで解釈しないと後ろの TZ で前月にずれる。 */
function formatPeriod(iso: string): string {
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return iso || '対象月不明';
  return `${d.getUTCFullYear()}年${d.getUTCMonth() + 1}月`;
}

/** ISO 文字列（UTC）から「YYYY/MM/DD」を作る。未設定・不正値は「—」。 */
function formatDate(iso?: string): string {
  if (!iso) return '—';
  const d = new Date(iso);
  if (Number.isNaN(d.getTime())) return '—';
  return `${d.getUTCFullYear()}/${String(d.getUTCMonth() + 1).padStart(2, '0')}/${String(d.getUTCDate()).padStart(2, '0')}`;
}
