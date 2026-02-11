import { MagnifyingGlassIcon } from '@heroicons/react/24/solid';

interface SearchBoxProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
}

export default function SearchBox({ value, onChange, placeholder = '検索' }: SearchBoxProps) {
  return (
    <div className="relative">
      <MagnifyingGlassIcon className="absolute left-4 top-1/2 h-5 w-5 -translate-y-1/2 text-gray-400" />
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="w-full rounded-lg border border-gray-300 py-2.5 pl-12 pr-4 focus:border-primary-500 focus:ring-1 focus:ring-primary-500 focus:outline-none transition-colors duration-150 placeholder-gray-400"
      />
    </div>
  );
}
