import { useEffect, useState } from 'react';
import { useAiChat } from '../hooks/useAiChat';

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

  const getScoreColor = (score: number) => {
    if (score >= 8) return 'text-green-600';
    if (score >= 6) return 'text-yellow-600';
    if (score >= 4) return 'text-orange-600';
    return 'text-red-600';
  };

  const getBarColor = (score: number) => {
    if (score >= 8) return 'bg-green-500';
    if (score >= 6) return 'bg-yellow-500';
    if (score >= 4) return 'bg-orange-500';
    return 'bg-red-500';
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-500" />
      </div>
    );
  }

  if (history.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center h-64 text-slate-500">
        <p className="text-lg font-medium">スコア履歴がありません</p>
        <p className="text-sm mt-1">AIアシスタントでフィードバックを受けるとスコアが記録されます</p>
      </div>
    );
  }

  return (
    <div className="max-w-3xl mx-auto p-4 space-y-4">
      <div className="text-sm text-slate-500">
        全 {history.length} 件のフィードバック履歴
      </div>

      {history.map((item) => (
        <div
          key={item.sessionId}
          className="bg-white rounded-xl border border-slate-200 p-4"
        >
          <div className="flex items-center justify-between mb-3">
            <div>
              <h3 className="text-sm font-bold text-slate-700">
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
              <span className={`text-lg font-bold ${getScoreColor(item.overallScore)}`}>
                {item.overallScore.toFixed(1)}
              </span>
              <span className="text-xs text-slate-400">/10</span>
            </div>
          </div>

          <div className="space-y-1.5">
            {item.scores.map((axisScore) => (
              <div key={axisScore.axis} className="flex items-center gap-2">
                <span className="text-xs text-slate-600 w-24 flex-shrink-0 truncate">
                  {axisScore.axis}
                </span>
                <div className="flex-1 bg-slate-100 rounded-full h-2">
                  <div
                    className={`h-2 rounded-full ${getBarColor(axisScore.score)}`}
                    style={{ width: `${axisScore.score * 10}%` }}
                  />
                </div>
                <span className="text-xs font-bold text-slate-700 w-5 text-right">
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
