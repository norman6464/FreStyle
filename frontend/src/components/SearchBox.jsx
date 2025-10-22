export default function SearchBox({ value, onChange, placeholder = '検索' }) {
  return (
    <input
      type="text"
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      className="w-full p-2 border border-gray-300 rounded"
    />
  );
}
