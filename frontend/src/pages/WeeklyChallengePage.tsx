import { useWeeklyChallenge } from '../hooks/useWeeklyChallenge';
import Loading from '../components/Loading';

export default function WeeklyChallengePage() {
  const { challenge, loading } = useWeeklyChallenge();

  if (loading) return <Loading message="チャレンジを読み込み中..." />;

  if (!challenge) {
    return (
      <div className="max-w-lg mx-auto px-4 py-6">
        <h1 className="text-xl font-bold mb-4">ウィークリーチャレンジ</h1>
        <p className="text-gray-400 text-center py-8">今週のチャレンジはまだ設定されていません</p>
      </div>
    );
  }

  const progress = Math.min((challenge.completedSessions / challenge.targetSessions) * 100, 100);

  return (
    <div className="max-w-lg mx-auto px-4 py-6 space-y-4">
      <h1 className="text-xl font-bold">ウィークリーチャレンジ</h1>

      <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-5 space-y-4">
        <div className="flex items-center justify-between">
          <span className="text-xs text-gray-400">
            {new Date(challenge.weekStart).toLocaleDateString('ja-JP')} 〜 {new Date(challenge.weekEnd).toLocaleDateString('ja-JP')}
          </span>
          {challenge.isCompleted && (
            <span className="text-xs px-2 py-0.5 rounded-full bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400" data-testid="completed-badge">
              達成！
            </span>
          )}
        </div>

        <h2 className="text-lg font-bold">{challenge.title}</h2>
        <p className="text-sm text-gray-500 dark:text-gray-400">{challenge.description}</p>

        {/* プログレスバー */}
        <div className="space-y-1">
          <div className="flex justify-between text-sm">
            <span>{challenge.completedSessions} / {challenge.targetSessions} セッション</span>
            <span className="text-gray-400">{Math.round(progress)}%</span>
          </div>
          <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-3">
            <div
              className={`h-3 rounded-full transition-all ${
                challenge.isCompleted ? 'bg-green-500' : 'bg-blue-500'
              }`}
              style={{ width: `${progress}%` }}
              data-testid="progress-bar"
            />
          </div>
        </div>
      </div>
    </div>
  );
}
