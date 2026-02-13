import { useState } from 'react';

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
}

const PATTERNS = [
  { key: 'formal' as const, label: 'フォーマル', hint: '上司や顧客への報告・メールに' },
  { key: 'soft' as const, label: 'ソフト', hint: '指摘やお願いをする時に' },
  { key: 'concise' as const, label: '簡潔', hint: 'チャットやSlackで手短に' },
  { key: 'questioning' as const, label: '質問型', hint: '相手の意見を引き出したい時に' },
  { key: 'proposal' as const, label: '提案型', hint: '代替案を提示する時に' },
] as const;

export default function RephraseModal({ result, onClose }: RephraseModalProps) {
  const [copiedKey, setCopiedKey] = useState<string | null>(null);

  const handleCopy = async (text: string, key: string) => {
    await navigator.clipboard.writeText(text);
    setCopiedKey(key);
    setTimeout(() => setCopiedKey(null), 2000);
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-lg max-w-lg w-full mx-4 p-5">
        <h3 className="text-sm font-semibold text-slate-800 mb-4">言い換え提案</h3>

        {result === null ? (
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-slate-400 mr-3" />
            <span className="text-sm text-slate-500">言い換え中...</span>
          </div>
        ) : (
          <div className="space-y-3">
            {PATTERNS.map(({ key, label, hint }) => (
              <div key={key} className="border border-slate-200 rounded-lg p-3">
                <div className="flex items-center justify-between mb-1">
                  <div>
                    <span className="text-xs font-medium text-slate-600">{label}</span>
                    <span className="text-[10px] text-slate-400 ml-2">{hint}</span>
                  </div>
                  <button
                    onClick={() => handleCopy(result[key], key)}
                    className="text-xs text-slate-400 hover:text-slate-600 px-2 py-0.5 rounded hover:bg-slate-100 transition-colors"
                  >
                    {copiedKey === key ? 'コピーしました' : 'コピー'}
                  </button>
                </div>
                <p className="text-sm text-slate-700">{result[key]}</p>
              </div>
            ))}
          </div>
        )}

        <button
          onClick={onClose}
          className="w-full mt-4 py-2 text-sm text-slate-500 hover:text-slate-700 bg-primary-50 hover:bg-primary-100 rounded-lg transition-colors"
        >
          閉じる
        </button>
      </div>
    </div>
  );
}
