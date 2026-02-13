import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import ScenarioCard from '../components/ScenarioCard';
import { SkeletonCard } from '../components/Skeleton';
import type { PracticeScenario } from '../types';
import { usePractice } from '../hooks/usePractice';

const CATEGORIES = ['すべて', '顧客折衝', 'シニア・上司', 'チーム内'] as const;

export default function PracticePage() {
  const navigate = useNavigate();

  const [selectedCategory, setSelectedCategory] = useState<string>('すべて');
  const { scenarios, loading, fetchScenarios, createPracticeSession } = usePractice();

  useEffect(() => {
    fetchScenarios();
  }, [fetchScenarios]);

  const handleSelectScenario = async (scenario: PracticeScenario) => {
    const session = await createPracticeSession({ scenarioId: scenario.id });

    if (session) {
      navigate(`/chat/ask-ai/${session.id}`, {
        state: {
          sessionType: 'practice',
          scenarioId: scenario.id,
          scenarioName: scenario.name,
          initialPrompt: '練習開始',
        },
      });
    }
  };

  const filteredScenarios = selectedCategory === 'すべて'
    ? scenarios
    : scenarios.filter((s) => s.category === selectedCategory);

  return (
    <div className="p-6 max-w-2xl mx-auto">
      {/* ヘッダー */}
      <div className="mb-6">
        <h1 className="text-lg font-bold text-slate-800 mb-1">ビジネスシナリオ練習</h1>
        <p className="text-sm text-slate-500">
          AIが相手役を演じます。実践的なビジネスシーンでコミュニケーションスキルを磨きましょう。
        </p>
      </div>

      {/* カテゴリフィルター */}
      <div className="flex gap-2 mb-6 overflow-x-auto">
        {CATEGORIES.map((category) => (
          <button
            key={category}
            onClick={() => setSelectedCategory(category)}
            className={`px-3 py-1.5 rounded-full text-xs font-medium whitespace-nowrap transition-colors ${
              selectedCategory === category
                ? 'bg-primary-500 text-white'
                : 'bg-white text-slate-600 border border-slate-200 hover:bg-slate-50'
            }`}
          >
            {category}
          </button>
        ))}
      </div>

      {/* シナリオ一覧 */}
      {loading ? (
        <div className="grid gap-3">
          <SkeletonCard />
          <SkeletonCard />
          <SkeletonCard />
        </div>
      ) : filteredScenarios.length === 0 ? (
        <div className="text-center py-12 text-sm text-slate-500">
          シナリオがありません
        </div>
      ) : (
        <div className="grid gap-3">
          {filteredScenarios.map((scenario) => (
            <ScenarioCard
              key={scenario.id}
              scenario={scenario}
              onSelect={handleSelectScenario}
            />
          ))}
        </div>
      )}
    </div>
  );
}
