import { useNavigate } from 'react-router-dom';
import { ChatBubbleLeftIcon, UserPlusIcon } from '@heroicons/react/24/solid';
import ChatRepository from '../repositories/ChatRepository';

interface MemberItemProps {
  id: number;
  name: string;
  roomId?: number;
  email: string;
}

export default function MemberItem({ id, name, roomId, email }: MemberItemProps) {
  const navigate = useNavigate();

  const handleClick = async () => {
    try {
      if (roomId) {
        navigate(`/chat/users/${roomId}`);
        return;
      }

      const data = await ChatRepository.createRoom(id);
      if (data.roomId) {
        navigate(`/chat/users/${data.roomId}`);
      }
    } catch {
      // エラーはaxiosインターセプターが処理
    }
  };

  return (
    <div
      onClick={handleClick}
      className="bg-surface-1 rounded-2xl cursor-pointer overflow-hidden group border border-surface-3 hover:bg-surface-2 transition-colors duration-150 mb-3"
    >
      <div className="flex items-center p-4">
        <div className="w-14 h-14 bg-primary-500 rounded-full flex items-center justify-center text-white font-bold text-xl flex-shrink-0">
          {name?.charAt(0)?.toUpperCase() || '?'}
        </div>

        <div className="flex-1 ml-4 min-w-0">
          <p className="text-base font-bold text-[var(--color-text-primary)] truncate">
            {name || 'Unknown'}
          </p>
          <p className="text-sm text-[var(--color-text-muted)] truncate">
            {email}
          </p>
        </div>

        <div className="flex-shrink-0 ml-3">
          {roomId ? (
            <div className="bg-surface-2 text-primary-400 px-3 py-1.5 rounded-full flex items-center gap-1.5 group-hover:bg-surface-3 transition-colors">
              <ChatBubbleLeftIcon className="w-4 h-4" />
              <span className="text-xs font-semibold">チャット</span>
            </div>
          ) : (
            <div className="bg-green-900/30 text-green-400 px-3 py-1.5 rounded-full flex items-center gap-1.5 group-hover:bg-green-900/30 transition-colors">
              <UserPlusIcon className="w-4 h-4" />
              <span className="text-xs font-semibold">追加</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
