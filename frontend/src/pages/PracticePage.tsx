import ScenarioCard from '../components/ScenarioCard';
import { SkeletonCard } from '../components/Skeleton';
import FilterTabs from '../components/FilterTabs';
import DifficultyFilter from '../components/DifficultyFilter';
import SortSelector from '../components/SortSelector';
import FilterResetButton from '../components/FilterResetButton';
import SearchBox from '../components/SearchBox';
import { usePracticePage } from '../hooks/usePracticePage';

const CATEGORIES = ['すべて', 'ブックマーク', '顧客折衝', 'シニア・上司', 'チーム内'] as const;

export default function PracticePage() {
  const {
    selectedCategory,
    setSelectedCategory,
    selectedDifficulty,
    setSelectedDifficulty,
    selectedSort,
    setSelectedSort,
    searchQuery,
    setSearchQuery,
    isFilterActive,
    resetFilters,
    filteredScenarios,
    loading,
    handleSelectScenario,
    isBookmarked,
    toggleBookmark,
  } = usePracticePage();

  return (
    <div className="p-6 max-w-2xl mx-auto">
      {/* ヘッダー */}
      <div className="mb-5">
        <h1 className="text-sm font-semibold text-[var(--color-text-primary)] mb-1">ビジネスシナリオ練習</h1>
        <p className="text-xs text-[var(--color-text-muted)]">
          AIが相手役を演じます。実践的なビジネスシーンでコミュニケーションスキルを磨きましょう。
        </p>
      </div>

      {/* 検索ボックス */}
      <div className="mb-3">
        <SearchBox value={searchQuery} onChange={setSearchQuery} placeholder="シナリオを検索" />
      </div>

      {/* カテゴリタブ */}
      <FilterTabs
        tabs={[...CATEGORIES]}
        selected={selectedCategory}
        onSelect={setSelectedCategory}
        className="mb-3"
      />

      {/* 難易度フィルター・ソート・リセット */}
      <div className="flex items-center justify-between mb-5">
        <div className="flex items-center gap-3">
          <DifficultyFilter selected={selectedDifficulty} onChange={setSelectedDifficulty} />
          <FilterResetButton isActive={isFilterActive} onReset={resetFilters} />
        </div>
        <SortSelector selected={selectedSort} onChange={setSelectedSort} />
      </div>

      {/* シナリオ一覧 */}
      {loading ? (
        <div className="grid gap-3">
          <SkeletonCard />
          <SkeletonCard />
          <SkeletonCard />
        </div>
      ) : filteredScenarios.length === 0 ? (
        <div className="text-center py-12 text-xs text-[var(--color-text-muted)]">
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
