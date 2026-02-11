interface RephraseResult {
  formal: string;
  soft: string;
  concise: string;
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
      <div className="bg-white rounded-2xl shadow-xl max-w-lg w-full mx-4 p-6">
        <h3 className="text-lg font-bold text-gray-800 mb-4">言い換え提案</h3>

        {result === null ? (
          <div className="flex items-center justify-center py-8">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-500 mr-3" />
            <span className="text-gray-500">言い換え中...</span>
          </div>
        ) : (
          <div className="space-y-4">
            {([
              { key: 'formal' as const, label: 'フォーマル版', color: 'blue' },
              { key: 'soft' as const, label: 'ソフト版', color: 'green' },
              { key: 'concise' as const, label: '簡潔版', color: 'orange' },
            ]).map(({ key, label, color }) => (
              <div key={key} className="border border-gray-200 rounded-lg p-3">
                <div className="flex items-center justify-between mb-1">
                  <span className={`text-sm font-semibold text-${color}-600`}>{label}</span>
                  <button
                    onClick={() => handleCopy(result[key])}
                    className="text-xs text-gray-500 hover:text-primary-600 px-2 py-1 rounded hover:bg-gray-100 transition-colors"
                  >
                    コピー
                  </button>
                </div>
                <p className="text-sm text-gray-700">{result[key]}</p>
              </div>
            ))}
          </div>
        )}

        <button
          onClick={onClose}
          className="w-full mt-4 py-2 text-sm text-gray-500 hover:text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
        >
          閉じる
        </button>
      </div>
    </div>
  );
}
