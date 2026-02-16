import { useState } from 'react';
import { useFavoritePhrase } from '../hooks/useFavoritePhrase';
import SearchBox from '../components/SearchBox';
import FavoriteStatsCard from '../components/FavoriteStatsCard';
import EmptyState from '../components/EmptyState';
import ConfirmModal from '../components/ConfirmModal';
import { StarIcon, XMarkIcon } from '@heroicons/react/24/outline';
import type { FavoritePhrase } from '../types';

const PATTERN_FILTERS = ['すべて', 'フォーマル', 'ソフト', '簡潔'] as const;

export default function FavoritesPage() {
  const { phrases, filteredPhrases, searchQuery, setSearchQuery, patternFilter, setPatternFilter, removeFavorite } = useFavoritePhrase();
  const [deleteTarget, setDeleteTarget] = useState<FavoritePhrase | null>(null);

  return (
    <div className="max-w-3xl mx-auto p-4 space-y-3">
      <h2 className="text-sm font-semibold text-[var(--color-text-primary)]">お気に入りフレーズ</h2>

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
                    : 'bg-surface-3 text-[var(--color-text-tertiary)] hover:bg-surface-3'
                }`}
              >
                {f}
              </button>
            ))}
          </div>
        </>
      )}

      {phrases.length === 0 ? (
        <EmptyState
          icon={StarIcon}
          title="お気に入りフレーズがありません"
          description="言い換え提案で★をタップするとここに保存されます"
        />
      ) : filteredPhrases.length === 0 ? (
        <EmptyState
          icon={StarIcon}
          title="該当するフレーズがありません"
          description="検索条件やフィルターを変更してみてください"
        />
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
                  <span className="text-[10px] text-[var(--color-text-faint)]">
                    {new Date(phrase.createdAt).toLocaleDateString('ja-JP')}
                  </span>
                  <button
                    onClick={() => setDeleteTarget(phrase)}
                    aria-label="お気に入りから削除"
                    className="text-xs text-[var(--color-text-faint)] hover:text-rose-500 transition-colors"
                  >
                    <XMarkIcon className="w-4 h-4" />
                  </button>
                </div>
              </div>
              <p className="text-sm text-[var(--color-text-primary)] mb-1">{phrase.rephrasedText}</p>
              <p className="text-xs text-[var(--color-text-faint)]">元: {phrase.originalText}</p>
            </div>
          ))}
        </div>
      )}
      <ConfirmModal
        isOpen={deleteTarget !== null}
        title="フレーズを削除"
        message={`「${deleteTarget?.rephrasedText ?? ''}」を削除しますか？`}
        onConfirm={() => {
          if (deleteTarget) {
            removeFavorite(deleteTarget.id);
          }
          setDeleteTarget(null);
        }}
        onCancel={() => setDeleteTarget(null)}
      />
    </div>
  );
}
