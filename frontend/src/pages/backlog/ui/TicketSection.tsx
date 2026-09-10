import type { ReactNode } from 'react';

export interface TicketSectionProps {
  title: string;
  /** 見出しの右に添える件数（0 も出す。出したくなければ渡さない）。 */
  count?: number;
  /** 見出しの右端に置く小物（保存状態など）。 */
  action?: ReactNode;
  children: ReactNode;
}

/**
 * チケットの詳細を構成する節の器。
 *
 * 見出しは `<section aria-label>` にしない。詳細パネルは狭い幅と広い幅の 2 通りが
 * 同時に DOM へ出るので、名前つきの領域が二重になり読み上げの検査に落ちる。
 */
export default function TicketSection({ title, count, action, children }: TicketSectionProps) {
  return (
    <div className="mb-4">
      <div className="mb-1 flex items-baseline gap-2">
        <div className="text-[10.5px] font-semibold uppercase tracking-wide text-[var(--color-text-muted)]">
          {title}
          {count !== undefined && <span className="ml-1 tabular-nums">{count}</span>}
        </div>
        {action && <div className="ml-auto">{action}</div>}
      </div>
      {children}
    </div>
  );
}
