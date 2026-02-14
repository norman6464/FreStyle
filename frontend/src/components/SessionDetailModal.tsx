interface AxisScore {
  axis: string;
  score: number;
  comment: string;
}

interface Session {
  sessionId: number;
  sessionTitle: string;
  overallScore: number;
  scores: AxisScore[];
  createdAt: string;
}

interface SessionDetailModalProps {
  session: Session;
  onClose: () => void;
}

export default function SessionDetailModal({ session, onClose }: SessionDetailModalProps) {
  return (
    <div
      data-testid="modal-overlay"
      className="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4"
      onClick={onClose}
    >
      <div
        className="bg-white rounded-xl shadow-xl max-w-md w-full max-h-[80vh] overflow-y-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-5">
          <div className="flex items-start justify-between mb-4">
            <div>
              <h2 className="text-sm font-bold text-slate-800">
                {session.sessionTitle || `セッション #${session.sessionId}`}
              </h2>
              <p className="text-xs text-slate-400 mt-0.5">
                {new Date(session.createdAt).toLocaleDateString('ja-JP', {
                  year: 'numeric',
                  month: 'long',
                  day: 'numeric',
                })}
              </p>
            </div>
            <div className="text-right">
              <p className="text-xs text-slate-500">総合スコア</p>
              <p className="text-2xl font-bold text-primary-600">
                {session.overallScore.toFixed(1)}
              </p>
            </div>
          </div>

          <div className="space-y-3">
            {session.scores.map((axisScore) => (
              <div key={axisScore.axis} className="bg-slate-50 rounded-lg p-3">
                <div className="flex items-center justify-between mb-1.5">
                  <span className="text-xs font-medium text-slate-700">
                    {axisScore.axis}
                  </span>
                  <span className="text-sm font-bold text-slate-800">
                    {axisScore.score.toFixed(1)}
                  </span>
                </div>
                <div className="w-full bg-slate-200 rounded-full h-1.5 mb-2">
                  <div
                    className="h-1.5 rounded-full bg-primary-500"
                    style={{ width: `${axisScore.score * 10}%` }}
                  />
                </div>
                {axisScore.comment && (
                  <p className="text-xs text-slate-500">{axisScore.comment}</p>
                )}
              </div>
            ))}
          </div>

          <button
            onClick={onClose}
            className="w-full mt-4 py-2 text-xs font-medium text-slate-600 bg-slate-100 rounded-lg hover:bg-slate-200 transition-colors"
          >
            閉じる
          </button>
        </div>
      </div>
    </div>
  );
}
