import { useEffect, useState } from 'react';
import { Navigate } from 'react-router-dom';
import { useDispatch, useSelector } from 'react-redux';
import { setAuthData } from '../store/authSlice';

export default function AppWrapper({ children }) {
  const [loading, setLoading] = useState(true);
  const dispatch = useDispatch();
  const auth = useSelector((state) => state.auth);
  const hasAuth = !!auth?.sub; // sub があればログイン済みとみなす

  useEffect(() => {
    if (!hasAuth) {
      fetch(`${import.meta.env.VITE_API_BASE_URL}/api/user/me`, {
        credentials: 'include',
      })
        .then((res) => {
          if (!res.ok) throw new Error('未ログイン');
          return res.json();
        })
        .then((data) => {
          dispatch(setAuthData(data));
        })
        .catch(() => {})
        .finally(() => setLoading(false));
    } else {
      setLoading(false); // Redux にすでにある場合はロード完了
    }
  }, []);

  if (loading) return <div>ロード中...</div>;

  return hasAuth ? children : <Navigate to="/login" />;
}
