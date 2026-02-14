import { MagnifyingGlassIcon, XMarkIcon } from '@heroicons/react/24/solid';

interface SearchBoxProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
}

export default function SearchBox({ value, onChange, placeholder = '検索' }: SearchBoxProps) {
  return (
    <div className="relative">
      <MagnifyingGlassIcon className="absolute left-4 top-1/2 h-5 w-5 -translate-y-1/2 text-[var(--color-text-faint)]" />
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="w-full rounded-lg border border-surface-3 py-2.5 pl-12 pr-9 focus:border-primary-500 focus:ring-1 focus:ring-primary-500 focus:outline-none transition-colors duration-150 placeholder-slate-400"
      />
      {value && (
        <button
          onClick={() => onChange('')}
          aria-label="検索をクリア"
          className="absolute right-3 top-1/2 -translate-y-1/2 text-[var(--color-text-faint)] hover:text-[var(--color-text-secondary)] transition-colors"
        >
          <XMarkIcon className="h-4 w-4" />
        </button>
      )}
    </div>
  );
}
