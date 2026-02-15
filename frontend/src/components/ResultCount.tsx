interface ResultCountProps {
  filteredCount: number;
  totalCount: number;
  isFilterActive: boolean;
}

export default function ResultCount({ filteredCount, totalCount, isFilterActive }: ResultCountProps) {
  return (
    <p className="text-[10px] text-[var(--color-text-muted)] mb-2">
      {isFilterActive ? `${filteredCount} / ${totalCount}件` : `${totalCount}件`}
    </p>
  );
}
