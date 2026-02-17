package com.example.FreStyle.service;

import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.repository.ChatRoomRepository;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class ChatRoomServiceTest {

    @Mock
    private ChatRoomRepository chatRoomRepository;

    @InjectMocks
    private ChatRoomService chatRoomService;

    @Test
    @DisplayName("findChatRoomById: 存在するルームを返す")
    void findChatRoomById_returnsRoom() {
        ChatRoom room = new ChatRoom();
        room.setId(1);
        when(chatRoomRepository.findById(1)).thenReturn(Optional.of(room));

        ChatRoom result = chatRoomService.findChatRoomById(1);

        assertEquals(1, result.getId());
        verify(chatRoomRepository).findById(1);
    }

    @Test
    @DisplayName("findChatRoomById: 存在しないルームで例外")
    void findChatRoomById_throwsWhenNotFound() {
        when(chatRoomRepository.findById(999)).thenReturn(Optional.empty());

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> chatRoomService.findChatRoomById(999));
        assertEquals("ルームが存在しません。", ex.getMessage());
    }

    @Test
    @DisplayName("findChatRoomById: リポジトリが例外をスローした場合メッセージ付きで伝搬する")
    void findChatRoomById_propagatesRepositoryException() {
        when(chatRoomRepository.findById(1))
                .thenThrow(new RuntimeException("DB接続エラー"));

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> chatRoomService.findChatRoomById(1));
        assertEquals("DB接続エラー", ex.getMessage());
    }

    @Test
    @DisplayName("findChatRoomById: 返却されたChatRoomオブジェクトが同一インスタンスである")
    void findChatRoomById_returnsSameInstance() {
        ChatRoom room = new ChatRoom();
        room.setId(42);
        when(chatRoomRepository.findById(42)).thenReturn(Optional.of(room));

        ChatRoom result = chatRoomService.findChatRoomById(42);

        assertEquals(42, result.getId());
        assertSame(room, result);
    }

    @Test
    @DisplayName("findChatRoomById: 異なるIDで存在しないルームはそれぞれ例外をスローする")
    void findChatRoomById_throwsForDifferentNonExistentIds() {
        when(chatRoomRepository.findById(100)).thenReturn(Optional.empty());
        when(chatRoomRepository.findById(200)).thenReturn(Optional.empty());

        RuntimeException ex1 = assertThrows(RuntimeException.class,
                () -> chatRoomService.findChatRoomById(100));
        assertEquals("ルームが存在しません。", ex1.getMessage());

        RuntimeException ex2 = assertThrows(RuntimeException.class,
                () -> chatRoomService.findChatRoomById(200));
        assertEquals("ルームが存在しません。", ex2.getMessage());
    }
}
