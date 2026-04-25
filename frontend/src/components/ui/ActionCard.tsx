import { ReactNode } from 'react';
import { ArrowRightIcon } from '@heroicons/react/24/outline';
import { Link } from 'react-router-dom';

type ActionCardBaseProps = {
  /** タイトル（1 行で短く） */
  title: string;
  /** 補足説明（1〜2 行） */
  description?: string;
  /** 左側のアイコン */
  icon?: ReactNode;
  /** 強調度 */
  emphasis?: 'primary' | 'secondary';
  /** カード右上に表示する小さなバッジ（例: 「初心者向け」） */
  badge?: string;
};

type ActionCardLinkProps = ActionCardBaseProps & {
  /** 遷移先 URL（React Router 経由） */
  to: string;
  onClick?: never;
};

type ActionCardButtonProps = ActionCardBaseProps & {
  /** クリック時のコールバック（リンク用途でない場合） */
  onClick: () => void;
  to?: never;
};

type ActionCardProps = ActionCardLinkProps | ActionCardButtonProps;

/**
 * 「次に何をすればよいか」を明示するカード型 CTA。
 * 初心者向けレイアウトで最も重要な導線パーツ。
 * Link か Button のいずれかとして使う（排他）。
 */
export default function ActionCard({
  title,
  description,
  icon,
  emphasis = 'secondary',
  badge,
  ...props
}: ActionCardProps) {
  const isPrimary = emphasis === 'primary';

  const containerClasses = `group relative flex w-full items-start gap-4 rounded-xl border p-4 text-left transition-all duration-150 focus:outline-none focus:ring-2 focus:ring-primary-400 focus:ring-offset-2 focus:ring-offset-[var(--color-surface)] ${
    isPrimary
      ? 'border-primary-400/40 bg-gradient-to-br from-primary-500/15 to-surface-1 hover:from-primary-500/25 hover:-translate-y-0.5 hover:shadow-lg'
      : 'border-surface-3 bg-surface-1 hover:border-primary-400/40 hover:-translate-y-0.5 hover:shadow-md'
  }`;

  const content = (
    <>
      {icon && (
        <span
          className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-lg ${
            isPrimary
              ? 'bg-primary-500 text-white'
              : 'bg-surface-2 text-primary-300'
          }`}
          aria-hidden="true"
        >
          {icon}
        </span>
      )}
      <div className="flex-1 min-w-0">
        <div className="flex items-center justify-between gap-2">
          <p className="text-base font-semibold text-[var(--color-text-primary)]">{title}</p>
          {badge && (
            <span className="shrink-0 rounded-full border border-primary-400/40 bg-primary-500/10 px-2 py-0.5 text-xs font-medium text-primary-300">
              {badge}
            </span>
          )}
        </div>
        {description && (
          <p className="mt-1 text-sm text-[var(--color-text-secondary)] leading-relaxed">
            {description}
          </p>
        )}
      </div>
      <ArrowRightIcon
        className="mt-2 h-4 w-4 shrink-0 text-[var(--color-text-muted)] transition-transform group-hover:translate-x-0.5 group-hover:text-primary-300"
        aria-hidden="true"
      />
    </>
  );

  if ('to' in props && props.to) {
    return (
      <Link to={props.to} className={containerClasses}>
        {content}
      </Link>
    );
  }

  return (
    <button type="button" onClick={props.onClick} className={containerClasses}>
      {content}
    </button>
  );
}
