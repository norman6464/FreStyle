import type { TicketStatusCategory } from '../model/types';

export interface TicketStatusPillProps {
  name: string;
  /** backend の状態マスタが持つ実際の色（hex）。トークン化しない — 状態ごとの色分けは
   * ユーザーが管理画面で選んだ値そのものを見せる（設計 Ⅴ「状態の枠の色は color 列をそのまま使う」）。 */
  color: string;
  category: TicketStatusCategory;
  /** 選べる状態であることを示す chevron を出す（一覧の行・読み取り専用では出さない）。 */
  showChevron?: boolean;
  className?: string;
}

/** 状態 1 件をピルで表示する（一覧の行・詳細パネル・管理表で共用）。 */
export default function TicketStatusPill({
  name,
  color,
  showChevron = false,
  className = '',
}: TicketStatusPillProps) {
  return (
    <span
      className={`inline-flex items-center gap-1 rounded border px-1.5 py-0.5 text-xs font-medium ${className}`}
      style={{ borderColor: color, color }}
    >
      {name}
      {showChevron && (
        <svg
          width="10"
          height="10"
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="3"
          strokeLinecap="round"
          aria-hidden="true"
        >
          <path d="m6 9 6 6 6-6" />
        </svg>
      )}
    </span>
  );
}
