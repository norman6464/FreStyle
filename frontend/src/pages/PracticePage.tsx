import ScenarioCard from '../components/ScenarioCard';
import { SkeletonCard } from '../components/Skeleton';
import FilterTabs from '../components/FilterTabs';
import DifficultyFilter from '../components/DifficultyFilter';
import SortSelector from '../components/SortSelector';
import FilterResetButton from '../components/FilterResetButton';
import ResultCount from '../components/ResultCount';
import SearchBox from '../components/SearchBox';
import EmptyState from '../components/EmptyState';
import { AcademicCapIcon } from '@heroicons/react/24/outline';
import { usePracticePage } from '../hooks/usePracticePage';
import { PRACTICE_CATEGORY_TABS } from '../constants/scenarioLabels';
import { PageIntro, StepIndicator, GuidedHint, GlossaryTerm } from '../components/ui';
import { GLOSSARY } from '../constants/glossary';

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
    totalCount,
    filteredCount,
    loading,
    handleSelectScenario,
    isBookmarked,
    toggleBookmark,
  } = usePracticePage();

  return (
    <div className="p-6 max-w-2xl mx-auto">
      <PageIntro
        icon={<AcademicCapIcon className="h-6 w-6" />}
        title="ビジネスシナリオ練習"
        description={
          <>
            AI がビジネスの相手役を演じます。{' '}
            <GlossaryTerm
              term={GLOSSARY.scenario.term}
              definition={GLOSSARY.scenario.definition}
            />{' '}
            を選んで会話を始めると、終了後に{' '}
            <GlossaryTerm
              term={GLOSSARY.fiveAxisScore.term}
              definition={GLOSSARY.fiveAxisScore.definition}
            />{' '}
            で結果を確認できます。
          </>
        }
      />

      {/* 初回訪問ヒント */}
      <div className="mb-4">
        <GuidedHint
          title="練習モードの流れ"
          storageKey="hint:practice:intro-v1"
        >
          3 ステップで完了します。まずは気になるシナリオを 1 つ選んでみましょう。
        </GuidedHint>
      </div>

      {/* 進行ステップ */}
      <div className="mb-5">
        <StepIndicator
          steps={[
            { label: 'シナリオを選ぶ', description: 'まずはここから' },
            { label: 'AI と会話する', description: 'ロールプレイ' },
            { label: 'スコアを確認', description: '5 軸で振り返り' },
          ]}
          currentStep={0}
        />
      </div>

      {/* 検索ボックス */}
      <div className="mb-3">
        <SearchBox value={searchQuery} onChange={setSearchQuery} placeholder="シナリオを検索" />
      </div>

      {/* カテゴリタブ */}
      <FilterTabs
        tabs={[...PRACTICE_CATEGORY_TABS]}
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

      {/* 結果件数 */}
      {!loading && (
        <ResultCount filteredCount={filteredCount} totalCount={totalCount} isFilterActive={isFilterActive} />
      )}

      {/* シナリオ一覧 */}
      {loading ? (
        <div className="grid gap-3">
          <SkeletonCard />
          <SkeletonCard />
          <SkeletonCard />
        </div>
      ) : filteredScenarios.length === 0 ? (
        <div className="py-12">
          <EmptyState
            icon={AcademicCapIcon}
            title={isFilterActive ? '該当するシナリオがありません' : 'シナリオがありません'}
            description={isFilterActive ? 'フィルター条件を変更してみてください' : 'シナリオが読み込まれていません'}
            action={isFilterActive ? { label: 'フィルターをリセット', onClick: resetFilters } : undefined}
          />
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
