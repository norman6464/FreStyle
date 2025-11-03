import { useEffect, useState } from 'react';
import { Navigate } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import { setAuthData } from '../store/authSlice';

export default function AppWrapper({ children }) {
  const [loading, setLoading] = useState(true);
  const [user, setUser] = useState(null);
  const dispatch = useDispatch();

  useEffect(() => {
    fetch(`${import.meta.env.VITE_API_BASE_URL}/api/user/me`, {
      credentials: 'include',
    })
      .then((res) => {
        if (!res.ok) throw new Error('未ログイン');
        return res.json();
      })
      .then((data) => {
        setUser(data);
        dispatch(setAuthData(data));
      })
      .catch(() => setUser(null))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div>ロード中...</div>;

  return user ? children : <Navigate to="/login" />;
}
