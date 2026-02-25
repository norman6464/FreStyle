import { useCallback, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import ChatRepository from '../repositories/ChatRepository';

export function useChatRoomCreation() {
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);

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
      setError('チャットルームの作成に失敗しました');
    }
  }, [navigate]);

  return { openChat, error };
}
