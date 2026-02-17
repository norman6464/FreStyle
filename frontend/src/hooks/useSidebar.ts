import { useState, useEffect, useCallback } from 'react';
import { useDispatch } from 'react-redux';
import { useNavigate } from 'react-router-dom';
import { clearAuth } from '../store/authSlice';
import ChatRepository from '../repositories/ChatRepository';
import AuthRepository from '../repositories/AuthRepository';
import { useToast } from './useToast';

export function useSidebar() {
  const dispatch = useDispatch();
  const navigate = useNavigate();
  const { showToast } = useToast();
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
      showToast('success', 'ログアウトしました');
      navigate('/login');
    } catch {
      // サイレントに処理
    }
  }, [dispatch, navigate, showToast]);

  return { totalUnread, handleLogout };
}
