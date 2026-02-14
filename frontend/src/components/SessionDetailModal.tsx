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
        className="bg-surface-1 rounded-xl shadow-xl max-w-md w-full max-h-[80vh] overflow-y-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-5">
          <div className="flex items-start justify-between mb-4">
            <div>
              <h2 className="text-sm font-bold text-[var(--color-text-primary)]">
                {session.sessionTitle || `セッション #${session.sessionId}`}
              </h2>
              <p className="text-xs text-[var(--color-text-faint)] mt-0.5">
                {new Date(session.createdAt).toLocaleDateString('ja-JP', {
                  year: 'numeric',
                  month: 'long',
                  day: 'numeric',
                })}
              </p>
            </div>
            <div className="text-right">
              <p className="text-xs text-[var(--color-text-muted)]">総合スコア</p>
              <p className="text-2xl font-bold text-primary-400">
                {session.overallScore.toFixed(1)}
              </p>
            </div>
          </div>

          <div className="space-y-3">
            {session.scores.map((axisScore) => (
              <div key={axisScore.axis} className="bg-surface-2 rounded-lg p-3">
                <div className="flex items-center justify-between mb-1.5">
                  <span className="text-xs font-medium text-[var(--color-text-secondary)]">
                    {axisScore.axis}
                  </span>
                  <span className="text-sm font-bold text-[var(--color-text-primary)]">
                    {axisScore.score.toFixed(1)}
                  </span>
                </div>
                <div className="w-full bg-surface-3 rounded-full h-1.5 mb-2">
                  <div
                    className="h-1.5 rounded-full bg-primary-500"
                    style={{ width: `${axisScore.score * 10}%` }}
                  />
                </div>
                {axisScore.comment && (
                  <p className="text-xs text-[var(--color-text-muted)]">{axisScore.comment}</p>
                )}
              </div>
            ))}
          </div>

          <button
            onClick={onClose}
            className="w-full mt-4 py-2 text-xs font-medium text-[var(--color-text-tertiary)] bg-surface-3 rounded-lg hover:bg-surface-3 transition-colors"
          >
            閉じる
          </button>
        </div>
      </div>
    </div>
  );
}
