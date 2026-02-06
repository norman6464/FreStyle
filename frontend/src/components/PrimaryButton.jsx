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
      className="w-full bg-primary-500 text-white font-medium py-2.5 rounded-lg hover:bg-primary-600 transition-colors duration-150 disabled:opacity-50 disabled:cursor-not-allowed"
    >
      {children}
    </button>
  );
}
