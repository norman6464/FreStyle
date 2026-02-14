import { useEffect } from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { useDispatch } from 'react-redux';
import { setAuthData } from '../store/authSlice';
import authRepository from '../repositories/AuthRepository';

export function useLoginCallback() {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();
  const dispatch = useDispatch();
  const code = searchParams.get('code');
  const error = searchParams.get('error');

  useEffect(() => {
    if (error) {
      alert('認証エラーが発生しました。' + error);
      navigate('/login');
      return;
    }

    if (code) {
      authRepository
        .callback(code)
        .then(() => {
          dispatch(setAuthData());
          navigate('/');
        })
        .catch(() => {
          alert('認証に失敗しました。');
          navigate('/login');
        });
    } else {
      navigate('/login');
    }
  }, [code, error, dispatch, navigate]);
}
