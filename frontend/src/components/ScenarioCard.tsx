import type { PracticeScenario } from '../types';
import Card from './Card';

interface ScenarioCardProps {
  scenario: PracticeScenario;
  onSelect: (scenario: PracticeScenario) => void;
  isBookmarked?: boolean;
  onToggleBookmark?: (scenarioId: number) => void;
}

const difficultyLabel: Record<string, string> = {
  beginner: '初級',
  intermediate: '中級',
  advanced: '上級',
};

const categoryLabel: Record<string, string> = {
  customer: '顧客折衝',
  senior: 'シニア・上司',
  team: 'チーム内',
};

const difficultyDescription: Record<string, string> = {
  beginner: '基本的な報連相',
  intermediate: '利害関係の調整',
  advanced: '複雑な交渉・説得',
};

const difficultyColor: Record<string, string> = {
  beginner: 'bg-surface-3 text-[var(--color-text-secondary)] border-surface-3',
  intermediate: 'bg-surface-3 text-[var(--color-text-secondary)] border-surface-3',
  advanced: 'bg-surface-3 text-[var(--color-text-secondary)] border-surface-3',
};

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
        <span className="text-xs text-[var(--color-text-muted)]">{categoryLabel[scenario.category] || scenario.category}</span>
        <div className="flex items-center gap-2">
          {onToggleBookmark && (
            <button
              onClick={handleBookmarkClick}
              title={isBookmarked ? 'ブックマーク解除' : 'ブックマーク'}
              className="p-0.5 hover:bg-surface-2 rounded transition-colors"
            >
              <svg className="w-4 h-4" fill={isBookmarked ? 'currentColor' : 'none'} stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 5a2 2 0 012-2h10a2 2 0 012 2v16l-7-3.5L5 21V5z" />
              </svg>
            </button>
          )}
          <span className={`text-xs px-2 py-0.5 rounded border ${difficultyColor[scenario.difficulty] || 'bg-surface-2 text-[var(--color-text-muted)] border-surface-3'}`}>
            {difficultyLabel[scenario.difficulty] || scenario.difficulty}
          </span>
        </div>
      </div>
      <h3 className="text-sm font-medium text-[var(--color-text-primary)] mb-1">{scenario.name}</h3>
      <p className="text-xs text-[var(--color-text-muted)] mb-2">{scenario.description}</p>
      <p className="text-[11px] text-[var(--color-text-muted)] mb-3">
        {difficultyDescription[scenario.difficulty] || ''} ・ 約5〜10分
      </p>
      <div className="flex items-center gap-1 text-xs text-[var(--color-text-muted)]">
        <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
        </svg>
        <span>相手役: {scenario.roleName}</span>
      </div>
    </Card>
  );
}
