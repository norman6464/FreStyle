interface AxisScoreBarProps {
  score: number;
  barColorClass?: string;
}

export default function AxisScoreBar({ score, barColorClass = 'bg-primary-500' }: AxisScoreBarProps) {
  return (
    <div role="progressbar" className="w-full bg-surface-3 rounded-full h-1.5">
      <div
        className={`h-1.5 rounded-full ${barColorClass}`}
        style={{ width: `${score * 10}%` }}
      />
    </div>
  );
}
