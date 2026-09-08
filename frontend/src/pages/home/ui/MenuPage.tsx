import { DocumentTextIcon } from '@heroicons/react/24/outline';

import FeatureSection from './FeatureSection';
import FeatureCard from './FeatureCard';

/**
 * ホーム画面。
 *
 * 全ユーザーに同じレイアウトを出す（ナレッジ）。
 */
export default function MenuPage() {
  return (
    <div className="px-4 sm:px-6 pt-8 pb-24 max-w-6xl mx-auto">
      {/* ウェルカムセクション（データ非依存・即時表示） */}
      <section className="mb-8">
        <p className="text-xs font-semibold text-brand-700 uppercase tracking-widest mb-1">
          ダッシュボード
        </p>
        <h1 className="text-3xl font-bold text-[var(--color-text-primary)]">
          FreStyle へようこそ
        </h1>
        <p className="mt-2 text-sm text-[var(--color-text-muted)]">
          ナレッジに学習メモを書き留め、いつでも振り返れます。
        </p>
      </section>

      <div className="flex flex-col lg:flex-row gap-8 items-start">
        {/* ── 左メインコンテンツ ── */}
        <div className="flex-1 min-w-0 space-y-8 w-full">
          <FeatureSection title="ツール">
            <FeatureCard
              to="/kb"
              icon={DocumentTextIcon}
              title="ナレッジ"
              description="学習メモを書き留め、いつでも振り返れます。"
              color="taupe"
            />
          </FeatureSection>
        </div>
      </div>
    </div>
  );
}
