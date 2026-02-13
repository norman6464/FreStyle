import { ReactNode } from 'react';

interface AuthLayoutProps {
  children: ReactNode;
}

export default function AuthLayout({ children }: AuthLayoutProps) {
  return (
    <div className="min-h-screen flex items-center justify-center bg-surface px-4 py-12">
      <div className="w-full max-w-md bg-white rounded-2xl shadow-sm border border-slate-200 border-t-4 border-t-primary-500 overflow-hidden">
        <div className="py-8 flex items-center justify-center">
          <h1 className="text-2xl font-bold text-slate-900 tracking-tight">FreStyle</h1>
        </div>
        <div className="p-8">{children}</div>
      </div>
    </div>
  );
}
