import { useNavigate } from 'react-router-dom';

interface Session {
  sessionId: number;
  sessionTitle: string;
  overallScore: number;
  createdAt: string;
}

interface RecentSessionsCardProps {
  sessions: Session[];
}

function scoreColor(score: number): string {
  if (score >= 8) return 'text-emerald-400';
  if (score >= 6) return 'text-amber-400';
  return 'text-rose-400';
}

export default function RecentSessionsCard({ sessions }: RecentSessionsCardProps) {
  const navigate = useNavigate();
  const recent = sessions.slice(0, 3);

  if (recent.length === 0) return null;

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4">
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-xs font-semibold text-[#F0F0F0]">直近のセッション</h3>
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
            className="flex items-center justify-between py-1.5 border-b border-surface-3 last:border-0"
          >
            <div className="flex-1 min-w-0 mr-3">
              <p className="text-sm text-[#D0D0D0] truncate">{session.sessionTitle}</p>
              <p className="text-[10px] text-[#666666]">
                {new Date(session.createdAt).toLocaleDateString('ja-JP')}
              </p>
            </div>
            <span className={`text-sm font-semibold ${scoreColor(session.overallScore)}`}>
              {session.overallScore}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
