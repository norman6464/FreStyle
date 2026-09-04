import {
  CodeBracketIcon,
  DocumentTextIcon,
  BookOpenIcon,
} from '@heroicons/react/24/outline';

import FeatureSection from './FeatureSection';
import FeatureCard from './FeatureCard';

/**
 * ホーム画面。
 *
 * 全ユーザーに同じレイアウトを出す（コース・コード演習・ノート）。
 */
export default function MenuPage() {
  return (
    <div className="px-4 sm:px-6 pt-8 pb-24 max-w-6xl mx-auto">
      {/* ウェルカムセクション（データ非依存・即時表示） */}
      <section className="mb-8">
        <p className="text-xs font-semibold text-brand-500 uppercase tracking-widest mb-1">
          ダッシュボード
        </p>
        <h1 className="text-3xl font-bold text-[var(--color-text-primary)]">
          FreStyle へようこそ
        </h1>
        <p className="mt-2 text-sm text-[var(--color-text-muted)]">
          コースや演習で学習を進め、AI チャットで疑問を解決しましょう。
        </p>
      </section>

      <div className="flex flex-col lg:flex-row gap-8 items-start">
        {/* ── 左メインコンテンツ ── */}
        <div className="flex-1 min-w-0 space-y-8 w-full">
          <FeatureSection title="学習">
            <FeatureCard
              to="/courses"
              icon={BookOpenIcon}
              title="コース"
              description="体系的なカリキュラムで段階的に学べます。"
              color="emerald"
              badge="おすすめ"
              techLogos={['git', 'go', 'docker', 'php']}
            />
            <FeatureCard
              to="/code-editor"
              icon={CodeBracketIcon}
              title="コード演習"
              description="実際にコードを書いて手を動かしながら学べます。"
              color="emerald"
              techLogos={['go', 'php', 'javascript', 'typescript']}
            />
          </FeatureSection>

          <FeatureSection title="ツール">
            <FeatureCard
              to="/notes"
              icon={DocumentTextIcon}
              title="ノート"
              description="学習メモを書き留め、いつでも振り返れます。"
              color="taupe"
            />
          </FeatureSection>
        </div>
      </div>
    </div>
  );
}
