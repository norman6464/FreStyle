import { ReactNode } from 'react';
import { Link } from 'react-router-dom';

interface LinkTextProps {
  to: string;
  children: ReactNode;
}

export default function LinkText({ to, children }: LinkTextProps) {
  return (
    <Link
      to={to}
      className="text-sm text-primary-500 hover:text-primary-400 font-medium transition-colors duration-150 hover:underline"
    >
      {children}
    </Link>
  );
}
