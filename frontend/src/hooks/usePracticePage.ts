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
  const { scenarios, loading, fetchScenarios, createPracticeSession } = usePractice();
  const { bookmarkedIds, toggleBookmark, isBookmarked } = useBookmark();

  useEffect(() => {
    fetchScenarios();
  }, [fetchScenarios]);

  const filteredScenarios = useMemo(() => {
    if (selectedCategory === 'すべて') return scenarios;
    if (selectedCategory === 'ブックマーク') {
      return scenarios.filter((s) => bookmarkedIds.includes(s.id));
    }
    return scenarios.filter((s) => s.category === CATEGORY_LABEL_TO_DB[selectedCategory]);
  }, [scenarios, selectedCategory, bookmarkedIds]);

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
    filteredScenarios,
    loading,
    handleSelectScenario,
    isBookmarked,
    toggleBookmark,
  };
}
