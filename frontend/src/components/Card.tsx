import { ReactNode, HTMLAttributes } from 'react';

interface CardProps extends Omit<HTMLAttributes<HTMLDivElement>, 'className'> {
  children: ReactNode;
  className?: string;
}

export default function Card({ children, className = '', ...props }: CardProps) {
  return (
    <div className={`bg-surface-1 rounded-lg border border-surface-3 p-4 ${className}`} {...props}>
      {children}
    </div>
  );
}
