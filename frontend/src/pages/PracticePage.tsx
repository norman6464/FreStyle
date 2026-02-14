import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import ScenarioCard from '../components/ScenarioCard';
import { SkeletonCard } from '../components/Skeleton';
import type { PracticeScenario } from '../types';
import { usePractice } from '../hooks/usePractice';
import { useBookmark } from '../hooks/useBookmark';

const CATEGORIES = ['すべて', 'ブックマーク', '顧客折衝', 'シニア・上司', 'チーム内'] as const;

const CATEGORY_LABEL_TO_DB: Record<string, string> = {
  '顧客折衝': 'customer',
  'シニア・上司': 'senior',
  'チーム内': 'team',
};

export default function PracticePage() {
  const navigate = useNavigate();

  const [selectedCategory, setSelectedCategory] = useState<string>('すべて');
  const { scenarios, loading, fetchScenarios, createPracticeSession } = usePractice();
  const { bookmarkedIds, toggleBookmark, isBookmarked } = useBookmark();

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
    : selectedCategory === 'ブックマーク'
    ? scenarios.filter((s) => bookmarkedIds.includes(s.id))
    : scenarios.filter((s) => s.category === CATEGORY_LABEL_TO_DB[selectedCategory]);

  return (
    <div className="p-6 max-w-2xl mx-auto">
      {/* ヘッダー */}
      <div className="mb-5">
        <h1 className="text-sm font-semibold text-[#F0F0F0] mb-1">ビジネスシナリオ練習</h1>
        <p className="text-xs text-[#888888]">
          AIが相手役を演じます。実践的なビジネスシーンでコミュニケーションスキルを磨きましょう。
        </p>
      </div>

      {/* カテゴリタブ */}
      <div className="flex gap-1 mb-5 border-b border-surface-3">
        {CATEGORIES.map((category) => (
          <button
            key={category}
            onClick={() => setSelectedCategory(category)}
            className={`px-3 py-2 text-xs font-medium transition-colors border-b-2 -mb-px ${
              selectedCategory === category
                ? 'border-primary-500 text-primary-400'
                : 'border-transparent text-[#888888] hover:text-[#D0D0D0]'
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
        <div className="text-center py-12 text-xs text-[#888888]">
          シナリオがありません
        </div>
      ) : (
        <div className="grid gap-3">
          {filteredScenarios.map((scenario) => (
            <ScenarioCard
              key={scenario.id}
              scenario={scenario}
              onSelect={handleSelectScenario}
              isBookmarked={isBookmarked(scenario.id)}
              onToggleBookmark={toggleBookmark}
            />
          ))}
        </div>
      )}
    </div>
  );
}
