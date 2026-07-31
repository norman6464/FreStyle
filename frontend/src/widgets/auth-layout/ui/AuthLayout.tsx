import { ReactNode } from 'react';
import { useDocumentMeta } from '@/shared/lib/hooks/useDocumentMeta';

interface AuthLayoutProps {
  children: ReactNode;
  title?: string;
  footer?: ReactNode;
  /** ページ上部に固定するヘッダー(公開ページの導線など)。省略時は従来どおり中央寄せのみ。 */
  header?: ReactNode;
}

export default function AuthLayout({ children, title, footer, header }: AuthLayoutProps) {
  // 認証フロー画面(ログイン/パスワード再設定)は検索結果に出す価値がないため noindex。
  useDocumentMeta({ robots: 'noindex, nofollow' });

  return (
    // body は overflow:hidden(AppShell の二重スクロールバー対策)のため、公開ページは
    // 自身がスクロールコンテナを持つ。内側の min-h-full で短いコンテンツは従来どおり中央寄せ。
    <div className="h-full overflow-y-auto bg-surface">
      <div className="min-h-full flex flex-col">
        {header}

        <div className="flex flex-1 flex-col items-center justify-center px-4 py-12">
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
      </div>
    </div>
  );
}
