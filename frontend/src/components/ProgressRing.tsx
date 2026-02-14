interface ProgressRingProps {
  value: number;
  max: number;
  label?: string;
  size?: number;
}

function getStrokeColor(percentage: number): string {
  if (percentage >= 80) return 'text-emerald-400';
  if (percentage >= 60) return 'text-amber-400';
  return 'text-rose-400';
}

export default function ProgressRing({ value, max, label, size = 64 }: ProgressRingProps) {
  const percentage = max > 0 ? Math.min((value / max) * 100, 100) : 0;
  const strokeWidth = 4;
  const radius = (size - strokeWidth) / 2;
  const circumference = 2 * Math.PI * radius;
  const offset = circumference - (percentage / 100) * circumference;

  return (
    <div
      role="progressbar"
      aria-valuenow={value}
      aria-valuemin={0}
      aria-valuemax={max}
      className="relative inline-flex items-center justify-center"
    >
      <svg width={size} height={size} className="-rotate-90">
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke="currentColor"
          strokeWidth={strokeWidth}
          className="text-surface-3"
        />
        <circle
          cx={size / 2}
          cy={size / 2}
          r={radius}
          fill="none"
          stroke="currentColor"
          strokeWidth={strokeWidth}
          strokeDasharray={circumference}
          strokeDashoffset={offset}
          strokeLinecap="round"
          className={`${getStrokeColor(percentage)} transition-all`}
        />
      </svg>
      {label && (
        <span className="absolute text-xs font-semibold text-[var(--color-text-primary)]">
          {label}
        </span>
      )}
    </div>
  );
}
