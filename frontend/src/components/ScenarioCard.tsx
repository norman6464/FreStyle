import type { PracticeScenario } from '../types';
import Card from './Card';
import { BookmarkIcon as BookmarkOutlineIcon, UserIcon } from '@heroicons/react/24/outline';
import { BookmarkIcon as BookmarkSolidIcon } from '@heroicons/react/24/solid';
import { DIFFICULTY_STYLES } from '../constants/difficultyStyles';
import { DIFFICULTY_LABEL, CATEGORY_LABEL, DIFFICULTY_DESCRIPTION } from '../constants/scenarioLabels';

interface ScenarioCardProps {
  scenario: PracticeScenario;
  onSelect: (scenario: PracticeScenario) => void;
  isBookmarked?: boolean;
  onToggleBookmark?: (scenarioId: number) => void;
}

function getDifficultyColor(difficulty: string): string {
  const label = DIFFICULTY_LABEL[difficulty];
  return label && DIFFICULTY_STYLES[label]
    ? DIFFICULTY_STYLES[label]
    : 'bg-surface-2 text-[var(--color-text-muted)]';
}

export default function ScenarioCard({ scenario, onSelect, isBookmarked, onToggleBookmark }: ScenarioCardProps) {
  const handleBookmarkClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    onToggleBookmark?.(scenario.id);
  };

  return (
    <Card
      onClick={() => onSelect(scenario)}
      className="cursor-pointer hover:bg-surface-2 transition-colors"
    >
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs text-[var(--color-text-muted)]">{CATEGORY_LABEL[scenario.category] || scenario.category}</span>
        <div className="flex items-center gap-2">
          {onToggleBookmark && (
            <button
              onClick={handleBookmarkClick}
              title={isBookmarked ? 'ブックマーク解除' : 'ブックマーク'}
              className="p-0.5 hover:bg-surface-2 rounded transition-colors"
            >
              {isBookmarked ? (
                <BookmarkSolidIcon className="w-4 h-4" />
              ) : (
                <BookmarkOutlineIcon className="w-4 h-4" />
              )}
            </button>
          )}
          <span className={`text-xs px-2 py-0.5 rounded-full ${getDifficultyColor(scenario.difficulty)}`}>
            {DIFFICULTY_LABEL[scenario.difficulty] || scenario.difficulty}
          </span>
        </div>
      </div>
      <h3 className="text-sm font-medium text-[var(--color-text-primary)] mb-1">{scenario.name}</h3>
      <p className="text-xs text-[var(--color-text-muted)] mb-2">{scenario.description}</p>
      <p className="text-[11px] text-[var(--color-text-muted)] mb-3">
        {DIFFICULTY_DESCRIPTION[scenario.difficulty] || ''} ・ 約5〜10分
      </p>
      <div className="flex items-center gap-1 text-xs text-[var(--color-text-muted)]">
        <UserIcon className="w-3.5 h-3.5" />
        <span>相手役: {scenario.roleName}</span>
      </div>
    </Card>
  );
}
