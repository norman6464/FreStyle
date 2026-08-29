import { ReactNode } from 'react';

export type ButtonVariant = 'primary' | 'secondary' | 'ghost' | 'danger';
export type ButtonSize = 'sm' | 'md' | 'lg';

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  size?: ButtonSize;
  loading?: boolean;
  fullWidth?: boolean;
  children: ReactNode;
}

const VARIANT: Record<ButtonVariant, string> = {
  // 白文字に対して brand-500 は 3.7:1 で読みやすさの基準（4.5:1）に届かない。600 から始める（5.2:1）。
  primary: 'bg-brand-600 hover:bg-brand-700 active:bg-brand-800 text-white shadow-sm hover:shadow',
  secondary:
    'border border-[var(--color-border-hover)] bg-surface-1 shadow-sm hover:bg-surface-2 hover:shadow active:bg-surface-3 text-[var(--color-text-secondary)]',
  ghost: 'hover:bg-surface-2 active:bg-surface-3 text-[var(--color-text-secondary)]',
  danger: 'bg-red-500 hover:bg-red-600 active:bg-red-700 text-white',
};

const SIZE: Record<ButtonSize, string> = {
  sm: 'px-3 py-1.5 text-sm rounded-lg',
  md: 'px-4 py-2.5 text-sm rounded-lg',
  lg: 'px-6 py-3 text-base rounded-xl',
};

/** Button はバリアント・サイズ・ローディング状態を統一管理するプリミティブ。 */
export default function Button({
  variant = 'primary',
  size = 'md',
  loading = false,
  fullWidth = false,
  children,
  disabled,
  className = '',
  type = 'button',
  ...props
}: ButtonProps) {
  return (
    <button
      type={type}
      aria-busy={loading || undefined}
      disabled={disabled || loading}
      className={[
        'font-medium transition-colors duration-150',
        'disabled:opacity-50 disabled:cursor-not-allowed',
        'flex items-center justify-center gap-2',
        'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand-400 focus-visible:ring-offset-2',
        VARIANT[variant],
        SIZE[size],
        fullWidth ? 'w-full' : '',
        className,
      ]
        .filter(Boolean)
        .join(' ')}
      {...props}
    >
      {loading && (
        <svg
          data-testid="loading-spinner"
          aria-hidden="true"
          className="animate-spin h-4 w-4 shrink-0"
          viewBox="0 0 24 24"
          fill="none"
        >
          <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" />
          <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4z" />
        </svg>
      )}
      {children}
    </button>
  );
}
