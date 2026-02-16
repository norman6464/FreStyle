import { useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import ChatRepository from '../repositories/ChatRepository';

export function useChatRoomCreation() {
  const navigate = useNavigate();

  const openChat = useCallback(async (userId: number, roomId?: number) => {
    if (roomId) {
      navigate(`/chat/users/${roomId}`);
      return;
    }

    try {
      const data = await ChatRepository.createRoom(userId);
      if (data.roomId) {
        navigate(`/chat/users/${data.roomId}`);
      }
    } catch {
      // エラーはaxiosインターセプターが処理
    }
  }, [navigate]);

  return { openChat };
}
