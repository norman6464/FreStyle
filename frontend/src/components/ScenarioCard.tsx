import type { PracticeScenario } from '../types';

interface ScenarioCardProps {
  scenario: PracticeScenario;
  onSelect: (scenario: PracticeScenario) => void;
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
  beginner: 'bg-emerald-50 text-emerald-700 border-emerald-200',
  intermediate: 'bg-amber-50 text-amber-700 border-amber-200',
  advanced: 'bg-rose-50 text-rose-700 border-rose-200',
};

export default function ScenarioCard({ scenario, onSelect }: ScenarioCardProps) {
  return (
    <div
      onClick={() => onSelect(scenario)}
      className="bg-white rounded-lg border border-slate-200 p-4 cursor-pointer hover:bg-primary-50 transition-colors"
    >
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs text-slate-500">{categoryLabel[scenario.category] || scenario.category}</span>
        <span className={`text-xs px-2 py-0.5 rounded border ${difficultyColor[scenario.difficulty] || 'bg-slate-50 text-slate-500 border-slate-200'}`}>
          {difficultyLabel[scenario.difficulty] || scenario.difficulty}
        </span>
      </div>
      <h3 className="text-sm font-medium text-slate-800 mb-1">{scenario.name}</h3>
      <p className="text-xs text-slate-500 mb-2">{scenario.description}</p>
      <p className="text-[11px] text-slate-400 mb-3">
        {difficultyDescription[scenario.difficulty] || ''} ・ 約5〜10分
      </p>
      <div className="flex items-center gap-1 text-xs text-slate-400">
        <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
        </svg>
        <span>相手役: {scenario.roleName}</span>
      </div>
    </div>
  );
}
