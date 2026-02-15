import { useState, useEffect, useMemo, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import type { PracticeScenario } from '../types';
import { usePractice } from './usePractice';
import { useBookmark } from './useBookmark';

const CATEGORY_LABEL_TO_DB: Record<string, string> = {
  '顧客折衝': 'customer',
  'シニア・上司': 'senior',
  'チーム内': 'team',
};

export function usePracticePage() {
  const navigate = useNavigate();
  const [selectedCategory, setSelectedCategory] = useState<string>('すべて');
  const [selectedDifficulty, setSelectedDifficulty] = useState<string | null>(null);
  const { scenarios, loading, fetchScenarios, createPracticeSession } = usePractice();
  const { bookmarkedIds, toggleBookmark, isBookmarked } = useBookmark();

  useEffect(() => {
    fetchScenarios();
  }, [fetchScenarios]);

  const filteredScenarios = useMemo(() => {
    let result = scenarios;

    if (selectedCategory === 'ブックマーク') {
      result = result.filter((s) => bookmarkedIds.includes(s.id));
    } else if (selectedCategory !== 'すべて') {
      result = result.filter((s) => s.category === CATEGORY_LABEL_TO_DB[selectedCategory]);
    }

    if (selectedDifficulty) {
      result = result.filter((s) => s.difficulty === selectedDifficulty);
    }

    return result;
  }, [scenarios, selectedCategory, selectedDifficulty, bookmarkedIds]);

  const handleSelectScenario = useCallback(async (scenario: PracticeScenario) => {
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
  }, [createPracticeSession, navigate]);

  return {
    selectedCategory,
    setSelectedCategory,
    selectedDifficulty,
    setSelectedDifficulty,
    filteredScenarios,
    loading,
    handleSelectScenario,
    isBookmarked,
    toggleBookmark,
  };
}
