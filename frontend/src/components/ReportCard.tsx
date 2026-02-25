import { ArrowTrendingUpIcon, ArrowTrendingDownIcon } from '@heroicons/react/24/outline';
import type { LearningReport } from '../types';

interface ReportCardProps {
  report: LearningReport;
}

export default function ReportCard({ report }: ReportCardProps) {
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
