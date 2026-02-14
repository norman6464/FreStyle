import { useNavigate } from 'react-router-dom';

interface PracticeReminderCardProps {
  lastPracticeDate: string | null;
}

function getDaysAgo(dateStr: string): number {
  const now = new Date();
  const then = new Date(dateStr);
  const diffMs = now.getTime() - then.getTime();
  return Math.floor(diffMs / (1000 * 60 * 60 * 24));
}

export default function PracticeReminderCard({ lastPracticeDate }: PracticeReminderCardProps) {
  const navigate = useNavigate();

  if (!lastPracticeDate) return null;

  const daysAgo = getDaysAgo(lastPracticeDate);

  return (
    <div className="bg-surface-1 rounded-lg border border-surface-3 p-4 flex items-center justify-between">
      <div>
        {daysAgo === 0 ? (
          <p className="text-sm font-medium text-emerald-400">今日も練習済み！</p>
        ) : (
          <>
            <p className="text-sm font-medium text-[var(--color-text-secondary)]">最後の練習: {daysAgo}日前</p>
            {daysAgo >= 3 && (
              <p className="text-xs text-amber-400 mt-0.5">そろそろ練習しませんか？</p>
            )}
          </>
        )}
      </div>
      {daysAgo > 0 && (
        <button
          onClick={() => navigate('/practice')}
          className="text-xs font-medium text-white bg-primary-500 px-3 py-1.5 rounded-lg hover:bg-primary-600 transition-colors"
        >
          練習する
        </button>
      )}
    </div>
  );
}
