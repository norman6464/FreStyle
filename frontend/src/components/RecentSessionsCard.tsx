import { useNavigate } from 'react-router-dom';
import type { ScoreHistoryItem } from '../types';
import { getScoreTextColor } from '../utils/scoreColor';
import Card from './Card';
import { formatDate } from '../utils/formatters';

interface RecentSessionsCardProps {
  sessions: ScoreHistoryItem[];
}

export default function RecentSessionsCard({ sessions }: RecentSessionsCardProps) {
  const navigate = useNavigate();
  const recent = sessions.slice(0, 3);

  if (recent.length === 0) return null;

  return (
    <Card>
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-xs font-semibold text-[var(--color-text-primary)]">直近のセッション</h3>
        <button
          onClick={() => navigate('/scores')}
          className="text-[10px] text-primary-400 hover:text-primary-300"
        >
          すべて見る
        </button>
      </div>
      <div className="space-y-2">
        {recent.map((session) => (
          <div
            key={session.sessionId}
            className="flex items-center justify-between py-1.5 border-b border-surface-3 last:border-0 cursor-pointer hover:bg-[var(--color-surface-2)] -mx-1 px-1 rounded transition-colors"
            onClick={() => navigate(`/chat/ask-ai/${session.sessionId}`)}
          >
            <div className="flex-1 min-w-0 mr-3">
              <p className="text-sm text-[var(--color-text-secondary)] truncate">{session.sessionTitle}</p>
              <p className="text-[10px] text-[var(--color-text-faint)]">
                {formatDate(session.createdAt)}
              </p>
            </div>
            <span className={`text-sm font-semibold ${getScoreTextColor(session.overallScore)}`}>
              {session.overallScore}
            </span>
          </div>
        ))}
      </div>
    </Card>
  );
}
