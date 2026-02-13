import { useEffect, useState } from 'react';
import { useAiChat } from '../hooks/useAiChat';
import { SkeletonCard } from '../components/Skeleton';

interface AxisScore {
  axis: string;
  score: number;
  comment: string;
}

interface ScoreHistoryItem {
  sessionId: number;
  sessionTitle: string;
  overallScore: number;
  scores: AxisScore[];
  createdAt: string;
}

export default function ScoreHistoryPage() {
  const [history, setHistory] = useState<ScoreHistoryItem[]>([]);
  const { fetchScoreHistory, loading } = useAiChat();

  useEffect(() => {
    const loadHistory = async () => {
      const data = await fetchScoreHistory();
      setHistory(data);
    };
    loadHistory();
  }, [fetchScoreHistory]);

  if (loading) {
    return (
      <div className="max-w-3xl mx-auto p-4 space-y-4">
        <div className="text-sm text-slate-500">スコア履歴を読み込み中...</div>
        <SkeletonCard />
        <SkeletonCard />
        <SkeletonCard />
      </div>
    );
  }

  if (history.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-slate-500">
        <p className="text-sm font-medium">スコア履歴がありません</p>
        <p className="text-xs mt-1">AIアシスタントでフィードバックを受けるとスコアが記録されます</p>
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto p-4 space-y-3">
      <p className="text-xs text-slate-500">
        全 {history.length} 件のフィードバック履歴
      </p>

      {history.map((item) => (
        <div
          key={item.sessionId}
          className="bg-white rounded-lg border border-slate-200 p-4"
        >
          <div className="flex items-center justify-between mb-3">
            <div>
              <h3 className="text-sm font-medium text-slate-800">
                {item.sessionTitle || `セッション #${item.sessionId}`}
              </h3>
              <p className="text-xs text-slate-400 mt-0.5">
                {new Date(item.createdAt).toLocaleDateString('ja-JP', {
                  year: 'numeric',
                  month: 'long',
                  day: 'numeric',
                })}
              </p>
            </div>
            <div className="flex items-center gap-1">
              <span className="text-xs text-slate-500">総合</span>
              <span className="text-lg font-semibold text-slate-800">
                {item.overallScore.toFixed(1)}
              </span>
              <span className="text-xs text-slate-400">/10</span>
            </div>
          </div>

          <div className="space-y-1.5">
            {item.scores.map((axisScore) => (
              <div key={axisScore.axis} className="flex items-center gap-2">
                <span className="text-xs text-slate-500 w-24 flex-shrink-0 truncate">
                  {axisScore.axis}
                </span>
                <div className="flex-1 bg-slate-100 rounded-full h-1.5">
                  <div
                    className="h-1.5 rounded-full bg-primary-500"
                    style={{ width: `${axisScore.score * 10}%` }}
                  />
                </div>
                <span className="text-xs text-slate-600 w-5 text-right">
                  {axisScore.score}
                </span>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}
