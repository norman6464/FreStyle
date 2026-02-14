import { ReactNode } from 'react';

type BadgeVariant = 'success' | 'warning' | 'danger' | 'info' | 'neutral';
type BadgeSize = 'sm' | 'md';

interface BadgeProps {
  children: ReactNode;
  variant?: BadgeVariant;
  size?: BadgeSize;
  className?: string;
}

const VARIANT_STYLES: Record<BadgeVariant, string> = {
  success: 'bg-emerald-900/30 text-emerald-400',
  warning: 'bg-amber-900/30 text-amber-400',
  danger: 'bg-rose-900/30 text-rose-400',
  info: 'bg-blue-900/30 text-blue-400',
  neutral: 'bg-surface-3 text-[var(--color-text-muted)]',
};

const SIZE_STYLES: Record<BadgeSize, string> = {
  sm: 'text-[10px] px-2 py-0.5',
  md: 'text-xs px-2.5 py-1',
};

export default function Badge({ children, variant = 'neutral', size = 'sm', className = '' }: BadgeProps) {
  return (
    <span className={`inline-flex items-center font-medium rounded-full ${VARIANT_STYLES[variant]} ${SIZE_STYLES[size]} ${className}`}>
      {children}
    </span>
  );
}
