import { ArrowDownTrayIcon } from '@heroicons/react/24/outline';
import type { ScoreHistoryItem } from '../types';

interface ExportScoreHistoryButtonProps {
  history: ScoreHistoryItem[];
}

function toCSV(history: ScoreHistoryItem[]): string {
  // 全セッションの軸名を収集（順序統一）
  const allAxes = Array.from(
    new Set(history.flatMap((h) => h.scores.map((s) => s.axis)))
  );

  const header = ['日時', 'セッション名', '総合スコア', ...allAxes];
  const rows = history.map((h) => {
    const date = new Date(h.createdAt).toLocaleDateString('ja-JP', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    });
    const axisScores = allAxes.map((axis) => {
      const found = h.scores.find((s) => s.axis === axis);
      return found ? found.score.toFixed(1) : '';
    });
    return [date, h.sessionTitle || '', h.overallScore.toFixed(1), ...axisScores];
  });

  const csvContent = [header, ...rows]
    .map((row) => row.map((cell) => `"${String(cell).replace(/"/g, '""')}"`).join(','))
    .join('\n');

  // BOM付きでExcel対応
  return '\uFEFF' + csvContent;
}

export default function ExportScoreHistoryButton({ history }: ExportScoreHistoryButtonProps) {
  const handleExport = () => {
    const csv = toCSV(history);
    const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);

    const link = document.createElement('a');
    link.href = url;
    link.download = `スコア履歴_${new Date().toISOString().slice(0, 10)}.csv`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

  return (
    <button
      onClick={handleExport}
      disabled={history.length === 0}
      aria-label="CSVエクスポート"
      className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-[var(--color-text-secondary)] bg-surface-2 rounded-lg hover:bg-surface-3 transition-colors disabled:opacity-40 disabled:cursor-not-allowed"
    >
      <ArrowDownTrayIcon className="w-3.5 h-3.5" />
      CSV
    </button>
  );
}
