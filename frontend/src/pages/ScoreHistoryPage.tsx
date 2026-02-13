import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAiChat } from '../hooks/useAiChat';
import { SkeletonCard } from '../components/Skeleton';
import SkillRadarChart from '../components/SkillRadarChart';
import PracticeCalendar from '../components/PracticeCalendar';
import SkillTrendChart from '../components/SkillTrendChart';

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

const FILTERS = ['すべて', '練習', 'フリー'] as const;

function isPracticeSession(title: string): boolean {
  return title.startsWith('練習:') || title.startsWith('練習：');
}

const AXIS_ADVICE: Record<string, string> = {
  '論理的構成力': '論理的構成力を伸ばすシナリオで練習しましょう',
  '配慮表現': '配慮表現を伸ばすシナリオで練習しましょう',
  '要約力': '要約力を伸ばすシナリオで練習しましょう',
  '提案力': '提案力を伸ばすシナリオで練習しましょう',
  '質問・傾聴力': '質問・傾聴力を伸ばすシナリオで練習しましょう',
};

export default function ScoreHistoryPage() {
  const [history, setHistory] = useState<ScoreHistoryItem[]>([]);
  const [filter, setFilter] = useState<string>('すべて');
  const { fetchScoreHistory, loading } = useAiChat();
  const navigate = useNavigate();

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

  const filteredHistory = history.filter((item) => {
    if (filter === '練習') return isPracticeSession(item.sessionTitle);
    if (filter === 'フリー') return !isPracticeSession(item.sessionTitle);
    return true;
  });

  const latestSession = history[history.length - 1];
  const weakestAxis = latestSession
    ? [...latestSession.scores].sort((a, b) => a.score - b.score)[0]
    : null;

  return (
    <div className="max-w-3xl mx-auto p-4 space-y-3">
      {/* 弱点ベースのおすすめ練習 */}
      {weakestAxis && (
        <div className="bg-primary-50 rounded-lg border border-primary-100 p-4">
          <p className="text-xs font-semibold text-primary-700 mb-1">おすすめ練習</p>
          <p className="text-xs text-primary-600 mb-2">
            {AXIS_ADVICE[weakestAxis.axis] || `${weakestAxis.axis}を伸ばすシナリオで練習しましょう`}
          </p>
          <button
            onClick={() => navigate('/practice')}
            className="text-xs font-medium text-primary-700 bg-white px-3 py-1.5 rounded-lg border border-primary-200 hover:bg-primary-100 transition-colors"
          >
            練習一覧を見る
          </button>
        </div>
      )}

      {/* スキルレーダーチャート */}
      {latestSession && latestSession.scores.length > 0 && (
        <div className="bg-white rounded-lg border border-slate-200 p-4 flex justify-center">
          <SkillRadarChart scores={latestSession.scores} title="最新のスキルバランス" />
        </div>
      )}

      {/* スコア推移グラフ */}
      <div className="bg-white rounded-lg border border-slate-200 p-4">
        <p className="text-xs font-medium text-slate-700 mb-3">スコア推移</p>
        <div className="flex items-end gap-1 h-16">
          {history.map((item) => (
            <div
              key={item.sessionId}
              data-testid="trend-bar"
              className="flex-1 bg-primary-500 rounded-t"
              style={{ height: `${item.overallScore * 10}%` }}
              title={`${item.sessionTitle}: ${item.overallScore}`}
            />
          ))}
        </div>
      </div>

      {/* スキル別推移 */}
      <SkillTrendChart history={history} />

      {/* 練習カレンダー */}
      <PracticeCalendar practiceDates={history.map(h => h.createdAt)} />

      {/* フィルタタブ */}
      <div className="flex gap-1 border-b border-slate-200">
        {FILTERS.map((f) => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`px-3 py-2 text-xs font-medium transition-colors border-b-2 -mb-px ${
              filter === f
                ? 'border-primary-500 text-primary-600'
                : 'border-transparent text-slate-500 hover:text-slate-700'
            }`}
          >
            {f}
          </button>
        ))}
      </div>

      <p className="text-xs text-slate-500">
        全 {filteredHistory.length} 件のフィードバック履歴
      </p>

      {filteredHistory.map((item, index) => {
        // 元のhistory配列でのindexを使ってスコア変動を計算
        const originalIndex = history.indexOf(item);
        const prevItem = originalIndex > 0 ? history[originalIndex - 1] : null;
        const delta = prevItem
          ? Math.round((item.overallScore - prevItem.overallScore) * 10) / 10
          : null;

        return (
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
              <div className="flex items-center gap-2">
                {delta !== null && delta !== 0 && (
                  <span className={`text-xs font-medium ${delta > 0 ? 'text-emerald-600' : 'text-rose-600'}`}>
                    {delta > 0 ? `+${delta.toFixed(1)}` : `\u2212${Math.abs(delta).toFixed(1)}`}
                  </span>
                )}
                <div className="flex items-center gap-1">
                  <span className="text-xs text-slate-500">総合</span>
                  <span className="text-lg font-semibold text-slate-800">
                    {item.overallScore.toFixed(1)}
                  </span>
                  <span className="text-xs text-slate-400">/10</span>
                </div>
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
        );
      })}
    </div>
  );
}
