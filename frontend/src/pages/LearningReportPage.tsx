import { useLearningReport } from '../hooks/useLearningReport';
import EmptyState from '../components/EmptyState';
import Loading from '../components/Loading';
import ReportCard from '../components/ReportCard';
import { DocumentChartBarIcon } from '@heroicons/react/24/outline';

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
