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
}
