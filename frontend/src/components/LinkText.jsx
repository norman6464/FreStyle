import { Link } from 'react-router-dom';

export default function LinkText({ to, children }) {
  return (
    <Link
      to={to}
      className="text-sm text-primary-500 hover:text-primary-600 font-medium transition-colors duration-150 hover:underline"
    >
      {children}
    </Link>
  );
}
