import { useState, ComponentType, SVGProps } from 'react';
import { useNavigate, useLocation, Link } from 'react-router-dom';
import { useSelector } from 'react-redux';
import {
  HomeIcon,
  SparklesIcon,
  AcademicCapIcon,
  ChartBarIcon,
  StarIcon,
  DocumentTextIcon,
  BellIcon,
  DocumentChartBarIcon,
  UserCircleIcon,
  ArrowLeftOnRectangleIcon,
  SunIcon,
  MoonIcon,
  CodeBracketIcon,
  BuildingOffice2Icon,
  ChevronDoubleLeftIcon,
} from '@heroicons/react/24/outline';
import Loading from '../Loading';
import { useSidebar } from '../../hooks/useSidebar';
import { useTheme } from '../../hooks/useTheme';
import type { RootState } from '../../store';

interface SubItem {
  label: string;
  to: string;
  matchPrefix?: string;
}

interface RailSection {
  id: string;
  icon: ComponentType<SVGProps<SVGSVGElement>>;
  label: string;
  to?: string;
  matchExact?: boolean;
  matchPrefix?: string;
  subItems?: SubItem[];
}

const mainSections: RailSection[] = [
  { id: 'home', icon: HomeIcon, label: 'ホーム', to: '/', matchExact: true },
  { id: 'ai', icon: SparklesIcon, label: 'AI', to: '/chat/ask-ai', matchPrefix: '/chat/ask-ai' },
  { id: 'practice', icon: AcademicCapIcon, label: '練習', to: '/practice', matchPrefix: '/practice' },
  { id: 'code', icon: CodeBracketIcon, label: 'コード', to: '/code-editor', matchPrefix: '/code-editor' },
  { id: 'scores', icon: ChartBarIcon, label: 'スコア', to: '/scores', matchExact: true },
  { id: 'favorites', icon: StarIcon, label: 'お気に入り', to: '/favorites', matchExact: true },
  { id: 'notes', icon: DocumentTextIcon, label: 'ノート', to: '/notes', matchPrefix: '/notes' },
  { id: 'notifications', icon: BellIcon, label: '通知', to: '/notifications', matchExact: true },
  { id: 'reports', icon: DocumentChartBarIcon, label: 'レポート', to: '/reports', matchExact: true },
];

const profileSection: RailSection = {
  id: 'profile', icon: UserCircleIcon, label: 'プロフィール', to: '/profile/me', matchExact: true,
};

const adminSection: RailSection = {
  id: 'admin',
  icon: BuildingOffice2Icon,
  label: '管理',
  matchPrefix: '/admin',
  subItems: [
    { label: '会社一覧', to: '/admin/companies', matchPrefix: '/admin/companies' },
    { label: 'シナリオ管理', to: '/admin/scenarios', matchPrefix: '/admin/scenarios' },
    { label: '招待管理', to: '/admin/invitations', matchPrefix: '/admin/invitations' },
  ],
};

interface SidebarProps {
  onNavigate?: () => void;
}

function isActiveSection(section: RailSection, pathname: string): boolean {
  if (section.matchExact) return pathname === section.to;
  if (section.matchPrefix) return pathname.startsWith(section.matchPrefix);
  if (section.to) return pathname === section.to;
  return false;
}

interface RailItemProps {
  section: RailSection;
  active: boolean;
  panelOpen: boolean;
  onClick: () => void;
}

function RailItem({ section, active, panelOpen, onClick }: RailItemProps) {
  const highlighted = active || panelOpen;
  return (
    <button
      onClick={onClick}
      title={section.label}
      className={`relative flex flex-col items-center gap-0.5 w-12 py-2 rounded-md transition-colors ${
        highlighted
          ? 'text-primary-300 bg-surface-3'
          : 'text-[var(--color-text-muted)] hover:bg-surface-2 hover:text-[var(--color-text-primary)]'
      }`}
    >
      {highlighted && (
        <span className="absolute left-0 top-1.5 bottom-1.5 w-0.5 bg-primary-400 rounded-r" />
      )}
      <section.icon className="w-5 h-5 flex-shrink-0" />
      <span className="text-[9px] leading-none text-center line-clamp-1 w-full px-0.5">
        {section.label}
      </span>
    </button>
  );
}

export default function Sidebar({ onNavigate }: SidebarProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const { handleLogout, loggingOut } = useSidebar();
  const { theme, toggleTheme } = useTheme();
  const isAdmin = useSelector((state: RootState) => state.auth.isAdmin);

  const [openSectionId, setOpenSectionId] = useState<string | null>(null);

  const allSections = [...mainSections, profileSection, ...(isAdmin ? [adminSection] : [])];
  const activePanelSection = allSections.find(
    (s) => s.id === openSectionId && s.subItems && s.subItems.length > 0,
  );

  const handleRailClick = (section: RailSection) => {
    if (section.subItems && section.subItems.length > 0) {
      setOpenSectionId((prev) => (prev === section.id ? null : section.id));
    } else if (section.to) {
      navigate(section.to);
      setOpenSectionId(null);
      onNavigate?.();
    }
  };

  return (
    <>
      {loggingOut && <Loading fullscreen message="ログアウト中..." />}
      <aside className="flex h-full bg-surface-1 border-r border-surface-3 flex-shrink-0">
        {/* ─── アイコンレール (常時表示・56px) ─── */}
        <div className="flex flex-col w-14 h-full py-2 items-center">
          {/* メインナビ */}
          <div className="flex-1 flex flex-col gap-0.5 overflow-y-auto items-center w-full px-1">
            {mainSections.map((section) => (
              <RailItem
                key={section.id}
                section={section}
                active={isActiveSection(section, location.pathname)}
                panelOpen={openSectionId === section.id}
                onClick={() => handleRailClick(section)}
              />
            ))}
          </div>

          <div className="w-10 border-t border-surface-3 my-1" />

          {/* プロフィール・管理・テーマ・ログアウト */}
          <div className="flex flex-col gap-0.5 items-center w-full px-1">
            <RailItem
              section={profileSection}
              active={isActiveSection(profileSection, location.pathname)}
              panelOpen={false}
              onClick={() => handleRailClick(profileSection)}
            />
            {isAdmin && (
              <RailItem
                section={adminSection}
                active={isActiveSection(adminSection, location.pathname)}
                panelOpen={openSectionId === adminSection.id}
                onClick={() => handleRailClick(adminSection)}
              />
            )}
            <button
              onClick={toggleTheme}
              title={theme === 'dark' ? 'ライトモード' : 'ダークモード'}
              className="flex flex-col items-center gap-0.5 w-12 py-2 rounded-md text-[var(--color-text-muted)] hover:bg-surface-2 hover:text-[var(--color-text-secondary)] transition-colors"
            >
              {theme === 'dark'
                ? <SunIcon className="w-5 h-5" />
                : <MoonIcon className="w-5 h-5" />
              }
              <span className="text-[9px] leading-none">
                {theme === 'dark' ? 'ライト' : 'ダーク'}
              </span>
            </button>
            <button
              onClick={() => { onNavigate?.(); handleLogout(); }}
              title="ログアウト"
              className="flex flex-col items-center gap-0.5 w-12 py-2 rounded-md text-[var(--color-text-muted)] hover:bg-red-900/30 hover:text-red-400 transition-colors"
            >
              <ArrowLeftOnRectangleIcon className="w-5 h-5" />
              <span className="text-[9px] leading-none">ログアウト</span>
            </button>
          </div>
        </div>

        {/* ─── セカンダリパネル (折りたたみ可能・192px) ─── */}
        {activePanelSection && (
          <div className="w-48 flex flex-col border-l border-surface-3 h-full">
            <div className="flex items-center justify-between px-3 py-3 border-b border-surface-3">
              <span className="text-sm font-semibold">{activePanelSection.label}</span>
              <button
                onClick={() => setOpenSectionId(null)}
                title="パネルを閉じる"
                className="p-1 rounded text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] hover:bg-surface-2 transition-colors"
              >
                <ChevronDoubleLeftIcon className="w-4 h-4" />
              </button>
            </div>
            <nav className="flex-1 px-2 py-2 space-y-0.5 overflow-y-auto">
              {activePanelSection.subItems?.map((sub) => {
                const subActive = sub.matchPrefix
                  ? location.pathname.startsWith(sub.matchPrefix)
                  : location.pathname === sub.to;
                return (
                  <Link
                    key={sub.to}
                    to={sub.to}
                    onClick={onNavigate}
                    className={`block px-3 py-2 rounded-md text-sm transition-colors ${
                      subActive
                        ? 'bg-surface-3 text-primary-300'
                        : 'text-[var(--color-text-tertiary)] hover:bg-surface-2 hover:text-[var(--color-text-primary)]'
                    }`}
                  >
                    {sub.label}
                  </Link>
                );
              })}
            </nav>
          </div>
        )}
      </aside>
    </>
  );
}
