import { CheckIcon } from '@heroicons/react/24/solid';

type Step = {
  /** ステップのタイトル（短く） */
  label: string;
  /** ステップの補足説明（省略可） */
  description?: string;
};

type StepIndicatorProps = {
  /** すべてのステップ */
  steps: Step[];
  /** 現在アクティブなステップ（0 始まり） */
  currentStep: number;
  /** 追加 className */
  className?: string;
};

/**
 * 多段階操作の現在位置を可視化するステップインジケーター。
 * 練習モードなど「シナリオ選択 → 会話 → スコア確認」のような導線で使う。
 */
export default function StepIndicator({ steps, currentStep, className = '' }: StepIndicatorProps) {
  return (
    <ol
      className={`flex items-center gap-2 ${className}`}
      aria-label="進行状況"
    >
      {steps.map((step, index) => {
        const isCompleted = index < currentStep;
        const isActive = index === currentStep;
        const stepNumber = index + 1;

        return (
          <li
            key={step.label}
            className="flex items-center gap-2 flex-1 min-w-0"
            aria-current={isActive ? 'step' : undefined}
          >
            <div className="flex items-center gap-2 min-w-0">
              <span
                className={`flex h-7 w-7 shrink-0 items-center justify-center rounded-full border text-xs font-semibold transition-colors ${
                  isCompleted
                    ? 'bg-primary-500 border-primary-500 text-white'
                    : isActive
                    ? 'bg-primary-500/20 border-primary-400 text-primary-300'
                    : 'bg-surface-2 border-surface-3 text-[var(--color-text-muted)]'
                }`}
                aria-hidden="true"
              >
                {isCompleted ? <CheckIcon className="h-3.5 w-3.5" /> : stepNumber}
              </span>
              <div className="min-w-0">
                <p
                  className={`text-sm font-medium truncate ${
                    isActive
                      ? 'text-[var(--color-text-primary)]'
                      : 'text-[var(--color-text-tertiary)]'
                  }`}
                >
                  {step.label}
                  <span className="sr-only">
                    {`（ステップ ${stepNumber} / ${steps.length}${
                      isCompleted ? '・完了' : isActive ? '・実行中' : ''
                    }）`}
                  </span>
                </p>
                {step.description && (
                  <p className="text-xs text-[var(--color-text-muted)] truncate">
                    {step.description}
                  </p>
                )}
              </div>
            </div>
            {index < steps.length - 1 && (
              <div
                className={`hidden sm:block h-0.5 flex-1 rounded ${
                  isCompleted ? 'bg-primary-500' : 'bg-surface-3'
                }`}
                aria-hidden="true"
              />
            )}
          </li>
        );
      })}
    </ol>
  );
}
