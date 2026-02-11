import { ReactNode } from 'react';

interface PrimaryButtonProps {
  children: ReactNode;
  onClick?: () => void;
  type?: 'button' | 'submit' | 'reset';
  disabled?: boolean;
}

export default function PrimaryButton({
  children,
  onClick,
  type = 'button',
  disabled,
}: PrimaryButtonProps) {
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
