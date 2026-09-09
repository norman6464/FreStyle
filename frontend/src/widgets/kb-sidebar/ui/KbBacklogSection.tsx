import { useLocation, useNavigate } from 'react-router-dom';
import type { KbSpace } from '@/entities/kb';
import KbSectionHeading from './KbSectionHeading';

export interface KbBacklogSectionProps {
  workspaceSlug: string;
  spaces: KbSpace[];
}

/**
 * サイドバーの「バックログ」節。スペースごとに 1 行（設計 Ⅵ）。
 *
 * KbSidebar が既に取得済みの spaces をそのまま使う（バックログ用に取得し直さない）。
 * チケットを有効化していないスペースも含めて全部出す —
 * 「バックログを見て初めて有効化する」導線になる（設計 Ⅹ・既定 a）。
 */
export default function KbBacklogSection({ workspaceSlug, spaces }: KbBacklogSectionProps) {
  const navigate = useNavigate();
  const location = useLocation();

  if (spaces.length === 0) return null;

  return (
    <>
      <KbSectionHeading label="バックログ" divider />
      {spaces.map((space) => {
        const to = `/kb/backlog/${space.id}`;
        const active = location.pathname === to;
        return (
          <button
            key={space.id}
            type="button"
            onClick={() => navigate(to)}
            aria-current={active}
            // アクセシブルネームは「バックログ: <名前>」にする。visible text だけだと
            // 同じスペース名がナレッジの木の見出しにも出るため、role+name で一意に選べなくなる
            // （既存の KbSidebar.test.tsx が素のスペース名で問い合わせている）。
            aria-label={`バックログ: ${space.name}`}
            className={`flex w-full items-center gap-1.5 rounded-md px-2 py-1 text-left text-sm transition-colors ${
              active
                ? 'bg-surface-3 text-[var(--color-text-primary)]'
                : 'text-[var(--color-text-secondary)] hover:bg-surface-2'
            }`}
          >
            <svg
              width="13"
              height="13"
              viewBox="0 0 24 24"
              fill="none"
              stroke="currentColor"
              strokeWidth="2.2"
              strokeLinecap="round"
              aria-hidden="true"
              className="shrink-0"
            >
              <path d="m16 6 4 14" />
              <path d="M12 6v14" />
              <path d="M8 8v12" />
              <path d="M4 4v16" />
            </svg>
            <span className="truncate">{space.name}</span>
          </button>
        );
      })}
    </>
  );
}
