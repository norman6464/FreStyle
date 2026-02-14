import type { PracticeScenario } from '../types';

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
  beginner: 'bg-emerald-900/30 text-emerald-400 border-emerald-800',
  intermediate: 'bg-amber-900/30 text-amber-400 border-amber-800',
  advanced: 'bg-rose-900/30 text-rose-400 border-rose-800',
};

export default function ScenarioCard({ scenario, onSelect, isBookmarked, onToggleBookmark }: ScenarioCardProps) {
  const handleBookmarkClick = (e: React.MouseEvent) => {
    e.stopPropagation();
    onToggleBookmark?.(scenario.id);
  };

  return (
    <div
      onClick={() => onSelect(scenario)}
      className="bg-surface-1 rounded-lg border border-surface-3 p-4 cursor-pointer hover:bg-surface-2 transition-colors"
    >
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs text-[#888888]">{categoryLabel[scenario.category] || scenario.category}</span>
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
          <span className={`text-xs px-2 py-0.5 rounded border ${difficultyColor[scenario.difficulty] || 'bg-surface-2 text-[#888888] border-surface-3'}`}>
            {difficultyLabel[scenario.difficulty] || scenario.difficulty}
          </span>
        </div>
      </div>
      <h3 className="text-sm font-medium text-[#F0F0F0] mb-1">{scenario.name}</h3>
      <p className="text-xs text-[#888888] mb-2">{scenario.description}</p>
      <p className="text-[11px] text-[#666666] mb-3">
        {difficultyDescription[scenario.difficulty] || ''} ・ 約5〜10分
      </p>
      <div className="flex items-center gap-1 text-xs text-[#666666]">
        <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
        </svg>
        <span>相手役: {scenario.roleName}</span>
      </div>
    </div>
  );
}
