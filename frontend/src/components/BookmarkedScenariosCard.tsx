import { useNavigate } from 'react-router-dom';
import { useBookmarkedScenarios } from '../hooks/useBookmarkedScenarios';
import { useStartPracticeSession } from '../hooks/useStartPracticeSession';
import Card from './Card';

export default function BookmarkedScenariosCard() {
  const navigate = useNavigate();
  const { scenarios, loading } = useBookmarkedScenarios(3);
  const { startSession, starting } = useStartPracticeSession();

  if (loading || scenarios.length === 0) return null;

  return (
    <Card>
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-xs font-semibold text-[var(--color-text-primary)]">ブックマーク済みシナリオ</h3>
        <button
          onClick={() => navigate('/practice')}
          className="text-[10px] text-primary-400 hover:text-primary-300"
        >
          すべて見る
        </button>
      </div>
      <div className="space-y-2">
        {scenarios.map((scenario) => (
          <div
            key={scenario.id}
            className="flex items-center justify-between py-1.5 border-b border-surface-3 last:border-0"
          >
            <div className="flex-1 min-w-0 mr-3">
              <p className="text-sm text-[var(--color-text-secondary)] truncate">{scenario.name}</p>
              <p className="text-[10px] text-[var(--color-text-faint)]">
                {scenario.category}・{scenario.difficulty}
              </p>
            </div>
            <button
              onClick={() => startSession(scenario)}
              disabled={starting}
              className="text-[10px] px-2 py-1 rounded bg-primary-600 text-white hover:bg-primary-500 disabled:opacity-50"
            >
              練習開始
            </button>
          </div>
        ))}
      </div>
    </Card>
  );
}
