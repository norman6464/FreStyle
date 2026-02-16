import { useNavigate } from 'react-router-dom';
import Card from './Card';
import { DIFFICULTY_LABEL, CATEGORY_LABEL } from '../constants/scenarioLabels';
import { DIFFICULTY_STYLES } from '../constants/difficultyStyles';

interface RecommendedScenarioCardProps {
  scenario: {
    id: number;
    name: string;
    description: string;
    category: string;
    roleName: string;
    difficulty: string;
  };
  weakAxis: string;
}

export default function RecommendedScenarioCard({ scenario, weakAxis }: RecommendedScenarioCardProps) {
  const navigate = useNavigate();
  const difficultyLabel = DIFFICULTY_LABEL[scenario.difficulty] || scenario.difficulty;
  const categoryLabel = CATEGORY_LABEL[scenario.category] || scenario.category;

  return (
    <Card>
      <div className="flex items-center justify-between mb-2">
        <p className="text-xs font-medium text-primary-300">
          {weakAxis}を伸ばすおすすめシナリオ
        </p>
        <span className={`text-[10px] font-medium px-2 py-0.5 rounded ${DIFFICULTY_STYLES[difficultyLabel] || ''}`}>
          {difficultyLabel}
        </span>
      </div>

      <p className="text-sm font-semibold text-[var(--color-text-primary)] mb-1">{scenario.name}</p>
      <p className="text-xs text-[var(--color-text-muted)] mb-2 line-clamp-2">{scenario.description}</p>

      <div className="flex items-center justify-between">
        <span className="text-[10px] text-[var(--color-text-faint)]">{categoryLabel}</span>
        <button
          onClick={() => navigate('/practice')}
          className="text-xs font-medium text-primary-400 hover:text-primary-300 transition-colors"
        >
          このシナリオで練習する
        </button>
      </div>
    </Card>
  );
}
