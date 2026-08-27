import { useState } from 'react';
import { UserCircleIcon } from '@heroicons/react/24/outline';

import ProfilePage from './ProfilePage';

/**
 * SettingsPage — `/settings` 配下の設定ページ。
 *
 * 左に設定カテゴリのサブメニュー、 右に選択中カテゴリのコンテンツ。
 */
type SectionId = 'profile';

interface Section {
  id: SectionId;
  label: string;
  icon: typeof UserCircleIcon;
}

export default function SettingsPage() {
  const sections: Section[] = [
    { id: 'profile', label: 'プロフィール', icon: UserCircleIcon },
  ];

  const [activeSection, setActiveSection] = useState<SectionId>('profile');

  return (
    <div className="flex h-full">
      {/* 左サブメニュー */}
      <aside className="w-56 flex-shrink-0 border-r border-surface-3 bg-surface-1 p-4 hidden md:block">
        <h1 className="text-lg font-bold text-[var(--color-text-primary)] mb-4">設定</h1>
        <nav className="space-y-0.5">
          {sections.map((section) => {
            const Icon = section.icon;
            const active = activeSection === section.id;
            return (
              <button
                key={section.id}
                type="button"
                onClick={() => setActiveSection(section.id)}
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
          <h1 className="text-lg font-bold text-[var(--color-text-primary)]">設定</h1>
        </div>

        {activeSection === 'profile' && <ProfilePage />}
      </div>
    </div>
  );
}
