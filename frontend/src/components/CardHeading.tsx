import type { ReactNode } from 'react';

interface CardHeadingProps {
  children: ReactNode;
}

export default function CardHeading({ children }: CardHeadingProps) {
  return (
    <p className="text-xs font-medium text-[var(--color-text-secondary)] mb-3">
      {children}
    </p>
  );
}
