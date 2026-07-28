import type { ReactNode } from 'react';

interface FeatureSectionProps {
  title: string;
  children: ReactNode;
}

/** ホームのメニューを見出し付きのカードグリッドでまとめるセクション（「学習」「ツール」等）。 */
export default function FeatureSection({ title, children }: FeatureSectionProps) {
  return (
    <section className="space-y-3">
      <h2 className="text-xs font-semibold text-[var(--color-text-muted)] uppercase tracking-widest">
        {title}
      </h2>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        {children}
      </div>
    </section>
  );
}
