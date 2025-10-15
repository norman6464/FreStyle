export default function PrimaryButton({
  children,
  onClick,
  type = 'button',
  disabled,
}) {
  return (
    <button
      type={type}
      onClick={onClick}
      disabled={disabled}
      className="w-full bg-blue-600 text-white font-semibold py-2 rounded hover:bg-blue-700 disabled:opacity-50"
    >
      {children}
    </button>
  );
}
