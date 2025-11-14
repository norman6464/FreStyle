import { MagnifyingGlassIcon } from '@heroicons/react/24/solid';

export default function SearchBox({ value, onChange, placeholder = '検索' }) {
  return (
    <div className="relative">
      <MagnifyingGlassIcon className="absolute left-3 top-1/2 h-5 w-5 -translate-y-1/2 text-gray-400" />
      <input
        type="text"
        value={value}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        className="w-full rounded-lg border border-gray-300 py-3 pl-10 pr-4 focus:border-blue-500 focus:ring-blue-500 shadow-sm"
      />
    </div>
  );
}