import { Bars3Icon } from '@heroicons/react/24/outline';

interface TopBarProps {
  title: string;
  onMenuToggle: () => void;
}

export default function TopBar({ title, onMenuToggle }: TopBarProps) {
  // タイトル文字列は各 page の PageIntro が h1 として表示するため、
  // ここではビジュアルとして繰り返さず screen reader 用にだけ残す。
  // h-12 → h-10 に縮め、上部に空きすぎる帯ができていた問題も合わせて解消。
  return (
    <header className="h-10 bg-surface-1 border-b border-surface-3 flex items-center px-4 flex-shrink-0">
      <button
        onClick={onMenuToggle}
        className="md:hidden p-1.5 -ml-1.5 mr-2 text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)] hover:bg-surface-2 rounded-md transition-colors"
        aria-label="メニュー"
      >
        <Bars3Icon className="w-5 h-5" />
      </button>
      <h1 className="sr-only">{title}</h1>
    </header>
  );
}
