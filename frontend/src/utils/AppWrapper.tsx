import { ReactNode, useEffect, useState } from 'react';
import { Navigate } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import { setAuthData } from '../store/authSlice';

interface AppWrapperProps {
  children: ReactNode;
}

export default function AppWrapper({ children }: AppWrapperProps) {
  const [loading, setLoading] = useState(true);
  const [user, setUser] = useState<any>(null);
  const dispatch = useDispatch();

  useEffect(() => {
    fetch(`${import.meta.env.VITE_API_BASE_URL}/api/user/me`, {
      credentials: 'include',
    })
      .then((res) => {
        console.log('未ログインかまたは正しくトークンが渡されていない。');
        if (!res.ok) throw new Error('未ログイン');
        return res.json();
      })
      .then((data) => {
        setUser(data);
        // dispatch(setFlashMessage('ログインに成功しました.'));
        dispatch(setAuthData());
      })
      .catch(() => setUser(null))
      .finally(() => setLoading(false));
  }, []);

  if (loading) return <div>ロード中...</div>;

  return user ? children : <Navigate to="/login" />;
}
