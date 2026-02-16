import { useNavigate } from 'react-router-dom';
import type { ScoreHistoryItem } from '../types';
import AxisScoreBar from './AxisScoreBar';

interface SessionDetailModalProps {
  session: ScoreHistoryItem;
  onClose: () => void;
}

export default function SessionDetailModal({ session, onClose }: SessionDetailModalProps) {
  const navigate = useNavigate();

  return (
    <div
      data-testid="modal-overlay"
      className="fixed inset-0 bg-black/40 z-50 flex items-center justify-center p-4"
      onClick={onClose}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="session-detail-title"
        className="bg-surface-1 rounded-xl shadow-xl max-w-md w-full max-h-[80vh] overflow-y-auto"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="p-5">
          <div className="flex items-start justify-between mb-4">
            <div>
              <h2 id="session-detail-title" className="text-sm font-bold text-[var(--color-text-primary)]">
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
            {(Array.isArray(session.scores) ? session.scores : []).map((axisScore) => (
              <div key={axisScore.axis} className="bg-surface-2 rounded-lg p-3">
                <div className="flex items-center justify-between mb-1.5">
                  <span className="text-xs font-medium text-[var(--color-text-secondary)]">
                    {axisScore.axis}
                  </span>
                  <span className="text-sm font-bold text-[var(--color-text-primary)]">
                    {axisScore.score.toFixed(1)}
                  </span>
                </div>
                <div className="mb-2">
                  <AxisScoreBar score={axisScore.score} />
                </div>
                {axisScore.comment && (
                  <p className="text-xs text-[var(--color-text-muted)]">{axisScore.comment}</p>
                )}
              </div>
            ))}
          </div>

          <div className="flex gap-2 mt-4">
            <button
              onClick={() => navigate(`/chat/ask-ai/${session.sessionId}`)}
              className="flex-1 py-2 text-xs font-medium text-white bg-primary-600 rounded-lg hover:bg-primary-500 transition-colors"
            >
              会話を見る
            </button>
            <button
              onClick={onClose}
              className="flex-1 py-2 text-xs font-medium text-[var(--color-text-tertiary)] bg-surface-3 rounded-lg hover:bg-surface-3 transition-colors"
            >
              閉じる
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
