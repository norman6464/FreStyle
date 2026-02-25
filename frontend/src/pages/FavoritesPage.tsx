import { useState } from 'react';
import { useFavoritePhrase } from '../hooks/useFavoritePhrase';
import { useCopyToClipboard } from '../hooks/useCopyToClipboard';
import SearchBox from '../components/SearchBox';
import FavoriteStatsCard from '../components/FavoriteStatsCard';
import EmptyState from '../components/EmptyState';
import ConfirmModal from '../components/ConfirmModal';
import Loading from '../components/Loading';
import { StarIcon, XMarkIcon, ClipboardDocumentIcon, CheckIcon } from '@heroicons/react/24/outline';
import type { FavoritePhrase } from '../types';

const PATTERN_FILTERS = ['すべて', 'フォーマル', 'ソフト', '簡潔'] as const;

export default function FavoritesPage() {
  const { phrases, filteredPhrases, searchQuery, setSearchQuery, patternFilter, setPatternFilter, removeFavorite, loading } = useFavoritePhrase();
  const { copiedId, copyToClipboard } = useCopyToClipboard();
  const [deleteTarget, setDeleteTarget] = useState<FavoritePhrase | null>(null);

  if (loading) {
    return <Loading message="読み込み中..." className="min-h-[calc(100vh-3.5rem)]" />;
  }

  return (
    <div className="max-w-3xl mx-auto p-4 space-y-3">
      <h2 className="text-sm font-semibold text-[var(--color-text-primary)]">
        お気に入りフレーズ
        {phrases.length > 0 && (
          <span className="ml-2 text-xs font-normal text-[var(--color-text-muted)]">{phrases.length}件</span>
        )}
      </h2>

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
              <div className="flex items-start justify-between gap-2 mb-1">
                <p className="text-sm text-[var(--color-text-primary)] flex-1">{phrase.rephrasedText}</p>
                <button
                  onClick={() => copyToClipboard(phrase.id, phrase.rephrasedText)}
                  aria-label="フレーズをコピー"
                  className="flex-shrink-0 p-1 rounded hover:bg-surface-2 transition-colors"
                >
                  {copiedId === phrase.id ? (
                    <CheckIcon className="w-4 h-4 text-emerald-500" />
                  ) : (
                    <ClipboardDocumentIcon className="w-4 h-4 text-[var(--color-text-faint)]" />
                  )}
                </button>
              </div>
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
