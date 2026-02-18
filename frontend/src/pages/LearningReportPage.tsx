import { useLearningReport } from '../hooks/useLearningReport';
import EmptyState from '../components/EmptyState';
import Loading from '../components/Loading';
import { DocumentChartBarIcon, ArrowTrendingUpIcon, ArrowTrendingDownIcon } from '@heroicons/react/24/outline';
import type { LearningReport } from '../types';

function ReportCard({ report }: { report: LearningReport }) {
  const scoreChangeColor = (report.scoreChange ?? 0) >= 0 ? 'text-emerald-500' : 'text-rose-500';
  const ScoreChangeIcon = (report.scoreChange ?? 0) >= 0 ? ArrowTrendingUpIcon : ArrowTrendingDownIcon;

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-sm font-semibold text-[var(--color-text-primary)]">
          {report.year}年{report.month}月
        </h3>
        {report.scoreChange != null && (
          <div className={`flex items-center gap-1 text-xs font-medium ${scoreChangeColor}`}>
            <ScoreChangeIcon className="w-3.5 h-3.5" />
            {report.scoreChange > 0 ? '+' : ''}{report.scoreChange.toFixed(1)}
          </div>
        )}
      </div>

      <div className="grid grid-cols-3 gap-3 mb-3">
        <div className="text-center">
          <p className="text-lg font-bold text-primary-500">{report.totalSessions}</p>
          <p className="text-[10px] text-[var(--color-text-faint)]">セッション数</p>
        </div>
        <div className="text-center">
          <p className="text-lg font-bold text-[var(--color-text-primary)]">{report.averageScore.toFixed(1)}</p>
          <p className="text-[10px] text-[var(--color-text-faint)]">平均スコア</p>
        </div>
        <div className="text-center">
          <p className="text-lg font-bold text-[var(--color-text-secondary)]">{report.practiceDays}</p>
          <p className="text-[10px] text-[var(--color-text-faint)]">練習日数</p>
        </div>
      </div>

      {(report.bestAxis || report.worstAxis) && (
        <div className="flex gap-2">
          {report.bestAxis && (
            <span className="text-[10px] font-medium text-emerald-400 bg-emerald-900/20 px-2 py-0.5 rounded">
              強み: {report.bestAxis}
            </span>
          )}
          {report.worstAxis && (
            <span className="text-[10px] font-medium text-amber-400 bg-amber-900/20 px-2 py-0.5 rounded">
              課題: {report.worstAxis}
            </span>
          )}
        </div>
      )}
    </div>
  );
}

export default function LearningReportPage() {
  const { reports, loading, generating, generateReport } = useLearningReport();

  const handleGenerate = () => {
    const now = new Date();
    generateReport(now.getFullYear(), now.getMonth() + 1);
  };

  if (loading) {
    return (
      <div className="max-w-3xl mx-auto p-4">
        <Loading size="medium" message="レポートを読み込み中..." className="py-12" />
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto p-4 space-y-3">
      <div className="flex items-center justify-between">
        <h2 className="text-sm font-semibold text-[var(--color-text-primary)]">
          学習レポート
          {reports.length > 0 && (
            <span className="ml-2 text-xs font-normal text-[var(--color-text-muted)]">{reports.length}件</span>
          )}
        </h2>
        <button
          onClick={handleGenerate}
          disabled={generating}
          className="text-xs text-primary-500 hover:text-primary-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {generating ? '生成中...' : '今月のレポートを生成'}
        </button>
      </div>

      {reports.length === 0 ? (
        <EmptyState
          icon={DocumentChartBarIcon}
          title="レポートはありません"
          description="「今月のレポートを生成」ボタンで学習レポートを作成できます"
        />
      ) : (
        <div className="space-y-3">
          {reports.map((report) => (
            <ReportCard key={report.id} report={report} />
          ))}
        </div>
      )}
    </div>
  );
}
