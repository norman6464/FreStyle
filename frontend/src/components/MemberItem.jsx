export default function MemberItem({ name }) {
  return (
    <div className="flex items-center bg-white p-3 rounded shadow hover:bg-gray-100 transition">
      <div className="w-10 h-10 bg-blue-400 rounded-full flex items-center justify text-white font-bold mr-4">
        {name.charAt(4).toUpperCase()}
      </div>
      <span className="text-lg font-medium">{name}</span>
    </div>
  );
}
