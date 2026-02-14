import { useFavoritePhrase } from '../hooks/useFavoritePhrase';
import SearchBox from '../components/SearchBox';
import FavoriteStatsCard from '../components/FavoriteStatsCard';

const PATTERN_FILTERS = ['すべて', 'フォーマル', 'ソフト', '簡潔'] as const;

export default function FavoritesPage() {
  const { phrases, filteredPhrases, searchQuery, setSearchQuery, patternFilter, setPatternFilter, removeFavorite } = useFavoritePhrase();

  return (
    <div className="max-w-3xl mx-auto p-4 space-y-3">
      <h2 className="text-sm font-semibold text-[#F0F0F0]">お気に入りフレーズ</h2>

      {phrases.length > 0 && (
        <FavoriteStatsCard phrases={phrases} />
      )}

      {phrases.length > 0 && (
        <>
          <SearchBox
            value={searchQuery}
            onChange={setSearchQuery}
            placeholder="フレーズを検索..."
          />

          <div className="flex gap-1">
            {PATTERN_FILTERS.map((f) => (
              <button
                key={f}
                onClick={() => setPatternFilter(f)}
                className={`px-3 py-1.5 text-xs font-medium rounded-full transition-colors ${
                  patternFilter === f
                    ? 'bg-primary-500 text-white'
                    : 'bg-surface-3 text-[#A0A0A0] hover:bg-surface-3'
                }`}
              >
                {f}
              </button>
            ))}
          </div>
        </>
      )}

      {phrases.length === 0 ? (
        <div className="flex flex-col items-center justify-center h-64 text-[#888888]">
          <p className="text-sm font-medium">お気に入りフレーズがありません</p>
          <p className="text-xs mt-1">言い換え提案で☆をタップするとここに保存されます</p>
        </div>
      ) : filteredPhrases.length === 0 ? (
        <div className="text-center py-8">
          <p className="text-sm text-[#888888]">該当するフレーズがありません</p>
        </div>
      ) : (
        <div className="space-y-2">
          {filteredPhrases.map((phrase) => (
            <div
              key={phrase.id}
              className="bg-surface-1 rounded-lg border border-surface-3 p-4"
            >
              <div className="flex items-start justify-between mb-2">
                <span className="text-[10px] font-medium text-primary-400 bg-surface-2 px-2 py-0.5 rounded">
                  {phrase.pattern}
                </span>
                <div className="flex items-center gap-2">
                  <span className="text-[10px] text-[#666666]">
                    {new Date(phrase.createdAt).toLocaleDateString('ja-JP')}
                  </span>
                  <button
                    onClick={() => removeFavorite(phrase.id)}
                    aria-label="お気に入りから削除"
                    className="text-xs text-[#666666] hover:text-rose-500 transition-colors"
                  >
                    ✕
                  </button>
                </div>
              </div>
              <p className="text-sm text-[#F0F0F0] mb-1">{phrase.rephrasedText}</p>
              <p className="text-xs text-[#666666]">元: {phrase.originalText}</p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
