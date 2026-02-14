import { useState } from 'react';
import { useFavoritePhrase } from '../hooks/useFavoritePhrase';

interface RephraseResult {
  formal: string;
  soft: string;
  concise: string;
  questioning: string;
  proposal: string;
}

interface RephraseModalProps {
  result: RephraseResult | null;
  onClose: () => void;
  originalText?: string;
}

const PATTERNS = [
  { key: 'formal' as const, label: 'フォーマル', hint: '上司や顧客への報告・メールに' },
  { key: 'soft' as const, label: 'ソフト', hint: '指摘やお願いをする時に' },
  { key: 'concise' as const, label: '簡潔', hint: 'チャットやSlackで手短に' },
  { key: 'questioning' as const, label: '質問型', hint: '相手の意見を引き出したい時に' },
  { key: 'proposal' as const, label: '提案型', hint: '代替案を提示する時に' },
] as const;

export default function RephraseModal({ result, onClose, originalText = '' }: RephraseModalProps) {
  const [copiedKey, setCopiedKey] = useState<string | null>(null);
  const { saveFavorite, isFavorite } = useFavoritePhrase();
  const [savedKeys, setSavedKeys] = useState<Set<string>>(new Set());

  const handleCopy = async (text: string, key: string) => {
    await navigator.clipboard.writeText(text);
    setCopiedKey(key);
    setTimeout(() => setCopiedKey(null), 2000);
  };

  const handleFavorite = (text: string, label: string, key: string) => {
    if (!isFavorite(text, label) && !savedKeys.has(key)) {
      saveFavorite(originalText, text, label);
      setSavedKeys((prev) => new Set(prev).add(key));
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-surface-1 rounded-lg shadow-lg max-w-lg w-full mx-4 p-5">
        <h3 className="text-sm font-semibold text-[#F0F0F0] mb-4">言い換え提案</h3>

        {result === null ? (
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-[#666666] mr-3" />
            <span className="text-sm text-[#888888]">言い換え中...</span>
          </div>
        ) : (
          <div className="space-y-3">
            {PATTERNS.map(({ key, label, hint }) => {
              const isSaved = isFavorite(result[key], label) || savedKeys.has(key);
              return (
                <div key={key} className="border border-surface-3 rounded-lg p-3">
                  <div className="flex items-center justify-between mb-1">
                    <div>
                      <span className="text-xs font-medium text-[#A0A0A0]">{label}</span>
                      <span className="text-[10px] text-[#666666] ml-2">{hint}</span>
                    </div>
                    <div className="flex items-center gap-1">
                      <button
                        onClick={() => handleFavorite(result[key], label, key)}
                        aria-label={isSaved ? 'お気に入り済み' : 'お気に入りに追加'}
                        className={`text-xs px-1.5 py-0.5 rounded transition-colors ${
                          isSaved
                            ? 'text-amber-500'
                            : 'text-[#555555] hover:text-amber-400'
                        }`}
                      >
                        {isSaved ? '★' : '☆'}
                      </button>
                      <button
                        onClick={() => handleCopy(result[key], key)}
                        className="text-xs text-[#666666] hover:text-[#A0A0A0] px-2 py-0.5 rounded hover:bg-surface-2 transition-colors"
                      >
                        {copiedKey === key ? 'コピーしました' : 'コピー'}
                      </button>
                    </div>
                  </div>
                  <p className="text-sm text-[#D0D0D0]">{result[key]}</p>
                </div>
              );
            })}
          </div>
        )}

        <button
          onClick={onClose}
          className="w-full mt-4 py-2 text-sm text-[#888888] hover:text-[#D0D0D0] bg-surface-2 hover:bg-surface-3 rounded-lg transition-colors"
        >
          閉じる
        </button>
      </div>
    </div>
  );
}
