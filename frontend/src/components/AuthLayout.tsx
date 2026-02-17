import { ReactNode } from 'react';

interface AuthLayoutProps {
  children: ReactNode;
  title?: string;
  footer?: ReactNode;
}

export default function AuthLayout({ children, title, footer }: AuthLayoutProps) {
  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-surface px-4 py-12">
      <div className="mb-6 flex flex-col items-center">
        <svg
          className="w-10 h-10 text-[var(--color-text-primary)] mb-3"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="2"
          strokeLinecap="round"
          strokeLinejoin="round"
          aria-hidden="true"
        >
          <polyline points="16 18 22 12 16 6" />
          <polyline points="8 6 2 12 8 18" />
        </svg>
        {title && (
          <h1 className="text-2xl font-bold text-[var(--color-text-primary)] tracking-tight">
            {title}
          </h1>
        )}
      </div>
      <div className="w-full max-w-sm bg-surface-1 rounded-xl border border-surface-3 p-6">
        {children}
      </div>
      {footer && (
        <div className="w-full max-w-sm bg-surface-1 rounded-xl border border-surface-3 p-4 text-center mt-4">
          {footer}
        </div>
      )}
    </div>
  );
}
