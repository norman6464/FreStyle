import { ReactNode } from 'react';
import HelpTooltip from './HelpTooltip';

type GlossaryTermProps = {
  /** 用語のラベル（本文中に表示される文字） */
  term: string;
  /** ツールチップで表示する説明 */
  definition: ReactNode;
  /** 追加 className */
  className?: string;
};

/**
 * 専門用語をインラインで説明する。用語の右にヘルプアイコンを出し、
 * クリックで定義を表示する。初心者に優しい「?」つきの用語として使う。
 *
 * 例: <GlossaryTerm term="5軸評価" definition="..." />
 */
export default function GlossaryTerm({ term, definition, className = '' }: GlossaryTermProps) {
  return (
    <span className={`inline-flex items-center gap-1 ${className}`}>
      <span className="underline decoration-dotted decoration-primary-300 underline-offset-2">
        {term}
      </span>
      <HelpTooltip label={`${term}の意味を表示`}>{definition}</HelpTooltip>
    </span>
  );
}
