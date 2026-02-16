import { describe, it, expect, vi, beforeEach } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useChatRoomCreation } from '../useChatRoomCreation';
import ChatRepository from '../../repositories/ChatRepository';

vi.mock('../../repositories/ChatRepository');

const mockNavigate = vi.fn();
vi.mock('react-router-dom', () => ({
  useNavigate: () => mockNavigate,
}));

const mockedRepo = vi.mocked(ChatRepository);

describe('useChatRoomCreation', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('既存のroomIdがある場合はそのままチャットページに遷移する', async () => {
    const { result } = renderHook(() => useChatRoomCreation());

    await act(async () => {
      await result.current.openChat(1, 100);
    });

    expect(mockedRepo.createRoom).not.toHaveBeenCalled();
    expect(mockNavigate).toHaveBeenCalledWith('/chat/users/100');
  });

  it('roomIdがない場合はルーム作成後に遷移する', async () => {
    mockedRepo.createRoom.mockResolvedValue({ roomId: 200 });

    const { result } = renderHook(() => useChatRoomCreation());

    await act(async () => {
      await result.current.openChat(1, undefined);
    });

    expect(mockedRepo.createRoom).toHaveBeenCalledWith(1);
    expect(mockNavigate).toHaveBeenCalledWith('/chat/users/200');
  });

  it('ルーム作成失敗時は遷移しない', async () => {
    mockedRepo.createRoom.mockRejectedValue(new Error('失敗'));

    const { result } = renderHook(() => useChatRoomCreation());

    await act(async () => {
      await result.current.openChat(1, undefined);
    });

    expect(mockNavigate).not.toHaveBeenCalled();
  });
});
