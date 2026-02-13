import { useState, useEffect, useCallback } from 'react';
import { useDispatch } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { clearAuth } from '../store/authSlice';
import ChatRepository from '../repositories/ChatRepository';
import AuthRepository from '../repositories/AuthRepository';

export function useSidebar() {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const [totalUnread, setTotalUnread] = useState(0);

  useEffect(() => {
    const fetchUnread = async () => {
      try {
        const chatUsers = await ChatRepository.fetchChatUsers();
        const unread = chatUsers.reduce(
          (sum, u) => sum + (u.unreadCount || 0), 0
        );
        setTotalUnread(unread);
      } catch {
        // サイレントに処理
      }
    };
    fetchUnread();
  }, []);

  const handleLogout = useCallback(async () => {
    try {
      await AuthRepository.logout();
      dispatch(clearAuth());
      navigate('/login');
    } catch (err) {
      console.error('ログアウトエラー:', err);
    }
  }, [dispatch, navigate]);

  return { totalUnread, handleLogout };
}
