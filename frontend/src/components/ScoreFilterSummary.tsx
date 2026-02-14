interface ScoreFilterSummaryProps {
  scores: number[];
}

export default function ScoreFilterSummary({ scores }: ScoreFilterSummaryProps) {
  if (scores.length === 0) return null;

  const avg = Math.round((scores.reduce((s, v) => s + v, 0) / scores.length) * 10) / 10;
  const max = Math.max(...scores);
  const min = Math.min(...scores);

  const items = [
    { label: '件数', value: String(scores.length) },
    { label: '平均', value: avg.toFixed(1) },
    { label: '最高', value: max.toFixed(1) },
    { label: '最低', value: min.toFixed(1) },
  ];

  return (
    <div className="grid grid-cols-4 gap-2">
      {items.map((item) => (
        <div key={item.label} className="bg-surface-1 rounded-lg border border-surface-3 p-3 text-center">
          <p className="text-[10px] text-[var(--color-text-muted)]">{item.label}</p>
          <p className="text-sm font-semibold text-[var(--color-text-primary)] mt-0.5">{item.value}</p>
        </div>
      ))}
    </div>
  );
}
