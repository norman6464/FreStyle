import { Link } from 'react-router-dom';

export default function LinkText({ to, children }) {
  return (
    <Link
      to={to}
      className="text-sm text-primary-600 hover:text-primary-700 font-semibold transition-colors duration-200 hover:underline"
    >
      {children}
    </Link>
  );
}
