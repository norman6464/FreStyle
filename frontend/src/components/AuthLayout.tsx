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
        <img
          src="/favicon.svg"
          alt=""
          aria-hidden="true"
          className="w-12 h-12 mb-3"
        />
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
