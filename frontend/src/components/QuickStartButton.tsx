import { useNavigate } from 'react-router-dom';
import { RocketLaunchIcon } from '@heroicons/react/24/solid';
import { useStartPracticeSession } from '../hooks/useStartPracticeSession';

interface QuickStartButtonProps {
  scenario: {
    id: number;
    name: string;
    description?: string;
    category?: string;
    roleName?: string;
    difficulty?: string;
  } | null;
}

export default function QuickStartButton({ scenario }: QuickStartButtonProps) {
  const navigate = useNavigate();
  const { startSession, starting } = useStartPracticeSession();

  const handleClick = () => {
    if (scenario) {
      startSession(scenario);
    } else {
      navigate('/practice');
    }
  };

  return (
    <button
      onClick={handleClick}
      disabled={starting}
      aria-label="練習を始める"
      className="w-full flex items-center gap-3 px-5 py-4 rounded-xl bg-gradient-to-r from-primary-600 to-primary-500 hover:from-primary-500 hover:to-primary-400 text-white shadow-lg shadow-primary-600/20 transition-all disabled:opacity-60"
    >
      <RocketLaunchIcon className="w-6 h-6 flex-shrink-0" />
      <div className="flex-1 text-left">
        <p className="text-sm font-bold">
          {starting ? '練習を開始中...' : '練習を始める'}
        </p>
        <p className="text-xs opacity-80 mt-0.5">
          {scenario ? scenario.name : 'シナリオを選んで練習'}
        </p>
      </div>
    </button>
  );
}
