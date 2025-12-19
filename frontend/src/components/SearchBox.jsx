import { MagnifyingGlassIcon } from '@heroicons/react/24/solid';

export default function SearchBox({ value, onChange, placeholder = '検索' }) {
  return (
    <div className="relative">
      <MagnifyingGlassIcon className="absolute left-4 top-1/2 h-5 w-5 -translate-y-1/2 text-gray-400" />
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="w-full rounded-lg border-2 border-gray-200 py-3 pl-12 pr-4 focus:border-primary-500 focus:ring-2 focus:ring-primary-100 focus:outline-none transition-all duration-200 shadow-sm placeholder-gray-400 font-medium"
      />
    </div>
  );
}
