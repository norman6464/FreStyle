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

const difficultyColor: Record<string, string> = {
  beginner: 'bg-green-100 text-green-700',
  intermediate: 'bg-yellow-100 text-yellow-700',
  advanced: 'bg-red-100 text-red-700',
};

export default function ScenarioCard({ scenario, onSelect }: ScenarioCardProps) {
  return (
    <div
      onClick={() => onSelect(scenario)}
      className="bg-white rounded-xl border border-slate-200 p-5 cursor-pointer hover:border-primary-300 hover:shadow-md transition-all duration-150"
    >
      <div className="flex items-center justify-between mb-2">
        <span className="text-xs font-medium text-slate-500">{scenario.category}</span>
        <span className={`text-xs font-medium px-2 py-0.5 rounded-full ${difficultyColor[scenario.difficulty] || 'bg-slate-100 text-slate-700'}`}>
          {difficultyLabel[scenario.difficulty] || scenario.difficulty}
        </span>
      </div>
      <h3 className="text-base font-bold text-slate-800 mb-1">{scenario.name}</h3>
      <p className="text-sm text-slate-600 mb-3">{scenario.description}</p>
      <div className="flex items-center gap-1 text-xs text-slate-500">
        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z" />
        </svg>
        <span>相手役: {scenario.roleName}</span>
      </div>
    </div>
  );
}
