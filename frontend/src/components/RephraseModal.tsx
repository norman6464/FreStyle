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

export default function RephraseModal({ result, onClose }: RephraseModalProps) {
  const handleCopy = (text: string) => {
    navigator.clipboard.writeText(text);
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
            {([
              { key: 'formal' as const, label: 'フォーマル' },
              { key: 'soft' as const, label: 'ソフト' },
              { key: 'concise' as const, label: '簡潔' },
              { key: 'questioning' as const, label: '質問型' },
              { key: 'proposal' as const, label: '提案型' },
            ]).map(({ key, label }) => (
              <div key={key} className="border border-slate-200 rounded-lg p-3">
                <div className="flex items-center justify-between mb-1">
                  <span className="text-xs font-medium text-slate-600">{label}</span>
                  <button
                    onClick={() => handleCopy(result[key])}
                    className="text-xs text-slate-400 hover:text-slate-600 px-2 py-0.5 rounded hover:bg-slate-100 transition-colors"
                  >
                    コピー
                  </button>
                </div>
                <p className="text-sm text-slate-700">{result[key]}</p>
              </div>
            ))}
          </div>
        )}

        <button
          onClick={onClose}
          className="w-full mt-4 py-2 text-sm text-slate-500 hover:text-slate-700 bg-slate-50 hover:bg-slate-100 rounded-lg transition-colors"
        >
          閉じる
        </button>
      </div>
    </div>
  );
}
