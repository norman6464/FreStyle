interface ProgressBarProps {
  percentage: number;
  barColorClass?: string;
}

export default function ProgressBar({ percentage, barColorClass = 'bg-primary-500' }: ProgressBarProps) {
  return (
    <div
      role="progressbar"
      aria-valuenow={percentage}
      aria-valuemin={0}
      aria-valuemax={100}
      className="w-full bg-surface-3 rounded-full h-2"
    >
      <div
        className={`h-2 rounded-full transition-all ${barColorClass}`}
        style={{ width: `${percentage}%` }}
      />
    </div>
  );
}
