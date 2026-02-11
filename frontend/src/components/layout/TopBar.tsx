import { Bars3Icon } from '@heroicons/react/24/outline';

interface TopBarProps {
  title: string;
  onMenuToggle: () => void;
}

export default function TopBar({ title, onMenuToggle }: TopBarProps) {
  return (
    <header className="h-12 bg-white border-b border-slate-200 flex items-center px-4 flex-shrink-0">
      <button
        onClick={onMenuToggle}
        className="md:hidden p-1.5 -ml-1.5 mr-2 text-slate-500 hover:text-slate-700 hover:bg-slate-100 rounded-md transition-colors"
        aria-label="メニュー"
      >
        <Bars3Icon className="w-5 h-5" />
      </button>
      <h1 className="text-sm font-semibold text-slate-800">{title}</h1>
    </header>
  );
}
