import { UserCircleIcon } from '@heroicons/react/24/outline';
import ProfilePage from './ProfilePage';

/**
 * SettingsPage — `/settings` 配下の設定ページ。
 *
 * 左に設定カテゴリのサブメニュー、 右に選択中カテゴリのコンテンツ。
 * 現状はカテゴリが「プロフィール」のみ。 将来「アカウント」「プライバシー」「通知設定」等を
 * 追加するときは、 SETTINGS_SECTIONS に追加して右側のコンテンツ条件分岐に追記する。
 */
const SETTINGS_SECTIONS = [
  {
    id: 'profile',
    label: 'プロフィール',
    icon: UserCircleIcon,
  },
  // 将来:
  // { id: 'account', label: 'アカウント', icon: ... },
  // { id: 'privacy', label: 'プライバシー', icon: ... },
] as const;

type SectionId = (typeof SETTINGS_SECTIONS)[number]['id'];

export default function SettingsPage() {
  // 現状はカテゴリが 1 つしかないので、 アクティブカテゴリは固定で「profile」 とする。
  // useState を入れる準備はしてあるので、 カテゴリ追加時は state を導入。
  const activeSection: SectionId = 'profile';

  return (
    <div className="flex h-full">
      {/* 左サブメニュー */}
      <aside className="w-56 flex-shrink-0 border-r border-surface-3 bg-surface-1 p-4 hidden md:block">
        <h1 className="text-lg font-bold text-[var(--color-text-primary)] mb-4">設定</h1>
        <nav className="space-y-0.5">
          {SETTINGS_SECTIONS.map((section) => {
            const Icon = section.icon;
            const active = activeSection === section.id;
            return (
              <button
                key={section.id}
                type="button"
                aria-current={active ? 'page' : undefined}
                className={`flex items-center gap-2 w-full px-2 py-2 rounded-md text-sm transition-colors ${
                  active
                    ? 'bg-[var(--color-nav-active)] text-[var(--color-text-primary)] font-medium'
                    : 'text-[var(--color-text-tertiary)] hover:bg-[var(--color-nav-hover)] hover:text-[var(--color-text-primary)]'
                }`}
              >
                <Icon className="w-4 h-4 flex-shrink-0" />
                <span>{section.label}</span>
              </button>
            );
          })}
        </nav>
      </aside>

      {/* 右コンテンツ */}
      <div className="flex-1 min-w-0 overflow-y-auto">
        {/* モバイル用ヘッダ（左サブメニューが隠れるとき） */}
        <div className="md:hidden border-b border-surface-3 px-4 py-3 bg-surface-1">
          <h1 className="text-lg font-bold text-[var(--color-text-primary)]">設定 / プロフィール</h1>
        </div>

        {activeSection === 'profile' && <ProfilePage />}
      </div>
    </div>
  );
}
