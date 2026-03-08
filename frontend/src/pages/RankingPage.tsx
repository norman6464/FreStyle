import { useRanking } from '../hooks/useRanking';
import Loading from '../components/Loading';
import { RankingEntry } from '../types';

function RankingEntryRow({ entry, isMe }: { entry: RankingEntry; isMe: boolean }) {
  const medalEmoji = entry.rank === 1 ? '\u{1F947}' : entry.rank === 2 ? '\u{1F948}' : entry.rank === 3 ? '\u{1F949}' : null;

  return (
    <div
      className={`flex items-center gap-3 p-3 rounded-lg ${
        isMe ? 'bg-blue-50 dark:bg-blue-900/20 ring-2 ring-blue-400' : 'bg-white dark:bg-gray-800'
      }`}
      data-testid={`ranking-entry-${entry.rank}`}
    >
      <div className="w-8 text-center font-bold text-lg">
        {medalEmoji ?? entry.rank}
      </div>
      <div className="w-9 h-9 rounded-full bg-gray-200 dark:bg-gray-700 overflow-hidden flex-shrink-0">
        {entry.iconUrl ? (
          <img src={entry.iconUrl} alt={entry.username} className="w-full h-full object-cover" />
        ) : (
          <div className="w-full h-full flex items-center justify-center text-sm text-gray-500">
            {entry.username.charAt(0)}
          </div>
        )}
      </div>
      <div className="flex-1 min-w-0">
        <p className="font-medium truncate">{entry.username}</p>
        <p className="text-xs text-gray-500">{entry.sessionCount}セッション</p>
      </div>
      <div className="text-right">
        <p className="font-bold text-lg">{entry.averageScore.toFixed(1)}</p>
        <p className="text-xs text-gray-500">平均スコア</p>
      </div>
    </div>
  );
}

export default function RankingPage() {
  const { ranking, period, changePeriod, loading, error } = useRanking();

  return (
    <div className="max-w-lg mx-auto px-4 py-6 space-y-4">
      <h1 className="text-xl font-bold">ランキング</h1>

      <div className="flex gap-2">
        <button
          onClick={() => changePeriod('weekly')}
          className={`px-4 py-2 rounded-full text-sm font-medium transition-colors ${
            period === 'weekly'
              ? 'bg-blue-500 text-white'
              : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300'
          }`}
          data-testid="period-weekly"
        >
          週間
        </button>
        <button
          onClick={() => changePeriod('monthly')}
          className={`px-4 py-2 rounded-full text-sm font-medium transition-colors ${
            period === 'monthly'
              ? 'bg-blue-500 text-white'
              : 'bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-300'
          }`}
          data-testid="period-monthly"
        >
          月間
        </button>
      </div>

      {loading && <Loading message="ランキングを読み込み中..." />}
      {error && <p className="text-red-500 text-center">{error}</p>}

      {ranking && !loading && (
        <>
          {ranking.myRanking && (
            <div className="space-y-1">
              <h2 className="text-sm font-semibold text-gray-500">あなたの順位</h2>
              <RankingEntryRow entry={ranking.myRanking} isMe={true} />
            </div>
          )}

          <div className="space-y-2">
            <h2 className="text-sm font-semibold text-gray-500">トップランキング</h2>
            {ranking.entries.length === 0 ? (
              <p className="text-gray-400 text-center py-8">まだランキングデータがありません</p>
            ) : (
              ranking.entries.map((entry) => (
                <RankingEntryRow
                  key={entry.userId}
                  entry={entry}
                  isMe={ranking.myRanking?.userId === entry.userId}
                />
              ))
            )}
          </div>
        </>
      )}
    </div>
  );
}
