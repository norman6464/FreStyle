import { Link } from 'react-router-dom';

export default function LinkText({ to, children }) {
  return (
    <Link to={to} className="text-sm text-blue-600 hover:underline">
      {children}
    </Link>
  );
}
