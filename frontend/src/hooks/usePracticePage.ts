import { useState, useEffect, useMemo, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import type { PracticeScenario } from '../types';
import { usePractice } from './usePractice';
import { useBookmark } from './useBookmark';
import { DIFFICULTY_ORDER } from '../constants/sortOptions';
import type { SortOption } from '../constants/sortOptions';
import { CATEGORY_LABEL_TO_DB } from '../constants/scenarioLabels';

export function usePracticePage() {
  const navigate = useNavigate();
  const [selectedCategory, setSelectedCategory] = useState<string>('すべて');
  const [selectedDifficulty, setSelectedDifficulty] = useState<string | null>(null);
  const [selectedSort, setSelectedSort] = useState<SortOption>('default');
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

    if (selectedSort === 'difficulty-asc') {
      result = [...result].sort((a, b) => (DIFFICULTY_ORDER[a.difficulty] ?? 0) - (DIFFICULTY_ORDER[b.difficulty] ?? 0));
    } else if (selectedSort === 'difficulty-desc') {
      result = [...result].sort((a, b) => (DIFFICULTY_ORDER[b.difficulty] ?? 0) - (DIFFICULTY_ORDER[a.difficulty] ?? 0));
    } else if (selectedSort === 'name') {
      result = [...result].sort((a, b) => a.name.localeCompare(b.name, 'ja'));
    }

    return result;
  }, [scenarios, selectedCategory, selectedDifficulty, selectedSort, bookmarkedIds]);

  const isFilterActive = selectedCategory !== 'すべて' || selectedDifficulty !== null || selectedSort !== 'default';

  const resetFilters = useCallback(() => {
    setSelectedCategory('すべて');
    setSelectedDifficulty(null);
    setSelectedSort('default');
  }, []);

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
    selectedSort,
    setSelectedSort,
    isFilterActive,
    resetFilters,
    filteredScenarios,
    loading,
    handleSelectScenario,
    isBookmarked,
    toggleBookmark,
  };
}
