import type { ComponentType, SVGProps } from 'react';
import { Link } from 'react-router-dom';
import { ArrowRightIcon } from '@heroicons/react/24/outline';
import LanguageIcon from '@/shared/ui/LanguageIcon';

type CardColor = 'brand' | 'emerald' | 'taupe' | 'blue';

interface FeatureCardProps {
  to: string;
  icon: ComponentType<SVGProps<SVGSVGElement>>;
  title: string;
  description: string;
  color: CardColor;
  badge?: string;
  /** 学べる技術のロゴ（Devicon）。指定時に説明の下へミニロゴ列を出す（FRESTYLE-179）。
      vendoring 済みの key（public/lang/*.svg）だけ渡すこと（未 vendoring は汎用アイコンにフォールバックし列が不揃いになる）。 */
  techLogos?: string[];
}

const iconBg: Record<CardColor, string> = {
  brand:   'bg-brand-100 text-brand-600',
  emerald: 'bg-emerald-100 text-emerald-700',
  taupe:   'bg-taupe-100 text-taupe-600',
  blue:    'bg-blue-100 text-blue-600',
};

/** ホームの機能カード 1 枚（コース / 演習 / AI チャット等へのリンク）。 */
export default function FeatureCard({ to, icon: Icon, title, description, color, badge, techLogos }: FeatureCardProps) {
  return (
    <Link
      to={to}
      className="group relative flex h-full flex-col p-5 rounded-xl border border-[var(--color-surface-3)] bg-[var(--color-surface-1)] shadow-sm hover:border-brand-300 hover:shadow-md hover:-translate-y-0.5 active:translate-y-0 active:shadow-sm transition-all duration-150"
    >
      {badge && (
        <span className="absolute top-4 right-4 text-[10px] font-semibold px-2 pt-px pb-[3px] rounded-full bg-emerald-100 text-emerald-700">
          {badge}
        </span>
      )}
      <div className="flex items-start gap-3">
        <div className={`w-9 h-9 rounded-lg flex items-center justify-center flex-shrink-0 ${iconBg[color]}`}>
          <Icon className="w-5 h-5 -translate-y-px" />
        </div>
        <div className="min-w-0 pt-1">
          <h3 className="font-semibold text-[var(--color-text-primary)] text-sm group-hover:text-brand-500 transition-colors">
            {title}
          </h3>
        </div>
      </div>
      <p className="mt-3 text-xs text-[var(--color-text-muted)] leading-relaxed">
        {description}
      </p>
      {/* 学べる技術のロゴ列（Devicon）。コース/演習カードにだけ付き、技術感を出す（FRESTYLE-179）。 */}
      {techLogos && techLogos.length > 0 && (
        <div className="mt-3 flex items-center gap-2" aria-hidden="true">
          {techLogos.map((tech) => (
            <LanguageIcon key={tech} language={tech} className="w-5 h-5" />
          ))}
        </div>
      )}
      <div className="mt-4 flex items-center gap-1 text-xs text-[var(--color-text-muted)] group-hover:text-brand-500 transition-colors">
        <span>開く</span>
        <ArrowRightIcon className="w-3 h-3 group-hover:translate-x-0.5 transition-transform" />
      </div>
    </Link>
  );
}
