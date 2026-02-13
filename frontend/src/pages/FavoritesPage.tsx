import { useFavoritePhrase } from '../hooks/useFavoritePhrase';

export default function FavoritesPage() {
  const { phrases, removeFavorite } = useFavoritePhrase();

  return (
    <div className="max-w-3xl mx-auto p-4 space-y-3">
      <h2 className="text-sm font-semibold text-slate-800">お気に入りフレーズ</h2>

      {phrases.length === 0 ? (
        <div className="flex flex-col items-center justify-center h-64 text-slate-500">
          <p className="text-sm font-medium">お気に入りフレーズがありません</p>
          <p className="text-xs mt-1">言い換え提案で☆をタップするとここに保存されます</p>
        </div>
      ) : (
        <div className="space-y-2">
          {phrases.map((phrase) => (
            <div
              key={phrase.id}
              className="bg-white rounded-lg border border-slate-200 p-4"
            >
              <div className="flex items-start justify-between mb-2">
                <span className="text-[10px] font-medium text-primary-600 bg-primary-50 px-2 py-0.5 rounded">
                  {phrase.pattern}
                </span>
                <div className="flex items-center gap-2">
                  <span className="text-[10px] text-slate-400">
                    {new Date(phrase.createdAt).toLocaleDateString('ja-JP')}
                  </span>
                  <button
                    onClick={() => removeFavorite(phrase.id)}
                    aria-label="お気に入りから削除"
                    className="text-xs text-slate-400 hover:text-rose-500 transition-colors"
                  >
                    ✕
                  </button>
                </div>
              </div>
              <p className="text-sm text-slate-800 mb-1">{phrase.rephrasedText}</p>
              <p className="text-xs text-slate-400">元: {phrase.originalText}</p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
