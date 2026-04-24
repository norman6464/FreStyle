import { ReactNode } from 'react';

type PageIntroProps = {
  /** ページ上部に表示する見出し */
  title: string;
  /** 1〜2行の補足説明。初心者に「この画面で何ができるか」を伝える */
  description?: ReactNode;
  /** タイトル横に置くアイコン */
  icon?: ReactNode;
  /** ページタイトルの右側に表示するアクション（ボタンなど） */
  actions?: ReactNode;
  /** h1 / h2 の切り替え。SEO上 1ページに h1 は 1 つなので必要に応じて調整 */
  headingLevel?: 1 | 2;
  /** 追加 className */
  className?: string;
};

/**
 * ページ上部の統一ヘッダー。全画面でトーンと情報量を揃えることで、
 * 初心者が「どのページでも同じパターンで情報が得られる」と学習できるようにする。
 */
export default function PageIntro({
  title,
  description,
  icon,
  actions,
  headingLevel = 1,
  className = '',
}: PageIntroProps) {
  const Heading = headingLevel === 1 ? 'h1' : 'h2';

  return (
    <header
      className={`mb-6 flex flex-col gap-3 border-b border-surface-3 pb-4 sm:flex-row sm:items-end sm:justify-between ${className}`}
    >
      <div className="min-w-0">
        <div className="flex items-center gap-2">
          {icon && <span className="text-primary-300 shrink-0">{icon}</span>}
          <Heading className="text-xl font-bold text-[var(--color-text-primary)] sm:text-2xl">
            {title}
          </Heading>
        </div>
        {description && (
          <p className="mt-1 text-sm text-[var(--color-text-secondary)] leading-relaxed">
            {description}
          </p>
        )}
      </div>
      {actions && <div className="flex flex-wrap items-center gap-2">{actions}</div>}
    </header>
  );
}
