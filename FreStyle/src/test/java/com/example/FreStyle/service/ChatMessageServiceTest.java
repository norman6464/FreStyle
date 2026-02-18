package com.example.FreStyle.service;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.ChatMessageDynamoRepository;
import com.example.FreStyle.repository.UserRepository;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.List;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class ChatMessageServiceTest {

    @Mock
    private ChatMessageDynamoRepository chatMessageDynamoRepository;

    @Mock
    private UserRepository userRepository;

    @InjectMocks
    private ChatMessageService chatMessageService;

    private User createUser(Integer id, String name) {
        User user = new User();
        user.setId(id);
        user.setName(name);
        return user;
    }

    @Test
    @DisplayName("getMessagesByRoom: メッセージ一覧をsenderName付きで返す")
    void getMessagesByRoom_returnsList() {
        ChatMessageDto msg1 = new ChatMessageDto("msg-1", 1, 10, null, "こんにちは", 1000L);
        ChatMessageDto msg2 = new ChatMessageDto("msg-2", 1, 10, null, "お元気ですか", 2000L);
        when(chatMessageDynamoRepository.findByRoomId(1)).thenReturn(List.of(msg1, msg2));

        User user = createUser(10, "テストユーザー");
        when(userRepository.findAllById(anySet())).thenReturn(List.of(user));

        List<ChatMessageDto> result = chatMessageService.getMessagesByRoom(1, 10);

        assertEquals(2, result.size());
        assertEquals("こんにちは", result.get(0).content());
        assertEquals("テストユーザー", result.get(0).senderName());
        assertEquals("お元気ですか", result.get(1).content());
    }

    @Test
    @DisplayName("addMessage: メッセージを保存してsenderName付きDTOを返す")
    void addMessage_savesAndReturnsDto() {
        ChatMessageDto saved = new ChatMessageDto("msg-100", 1, 10, null, "新メッセージ", 3000L);
        when(chatMessageDynamoRepository.save(1, 10, "新メッセージ")).thenReturn(saved);

        User sender = createUser(10, "送信者");
        when(userRepository.findById(10)).thenReturn(Optional.of(sender));

        ChatMessageDto dto = chatMessageService.addMessage(1, 10, "新メッセージ");

        assertEquals("msg-100", dto.id());
        assertEquals(1, dto.roomId());
        assertEquals(10, dto.senderId());
        assertEquals("送信者", dto.senderName());
        assertEquals("新メッセージ", dto.content());
    }

    @Test
    @DisplayName("deleteMessage: メッセージを削除する")
    void deleteMessage_deletesMessage() {
        chatMessageService.deleteMessage(1, 1000L);

        verify(chatMessageDynamoRepository).deleteByRoomIdAndCreatedAt(1, 1000L);
    }

    @Test
    @DisplayName("addMessage: DynamoDBリポジトリ保存失敗時に例外が伝搬する")
    void addMessage_propagatesRepositorySaveException() {
        when(chatMessageDynamoRepository.save(1, 10, "テスト"))
                .thenThrow(new RuntimeException("保存失敗"));

        assertThrows(RuntimeException.class,
                () -> chatMessageService.addMessage(1, 10, "テスト"));
    }

    @Test
    @DisplayName("getMessagesByRoom: 空のリストを返す")
    void getMessagesByRoom_returnsEmptyList() {
        when(chatMessageDynamoRepository.findByRoomId(1)).thenReturn(List.of());

        List<ChatMessageDto> result = chatMessageService.getMessagesByRoom(1, 10);

        assertTrue(result.isEmpty());
    }

    @Test
    @DisplayName("addMessage: ユーザーが見つからない場合senderNameはnull")
    void addMessage_nullSenderNameWhenUserNotFound() {
        ChatMessageDto saved = new ChatMessageDto("msg-1", 1, 999, null, "テスト", 4000L);
        when(chatMessageDynamoRepository.save(1, 999, "テスト")).thenReturn(saved);
        when(userRepository.findById(999)).thenReturn(Optional.empty());

        ChatMessageDto dto = chatMessageService.addMessage(1, 999, "テスト");

        assertNull(dto.senderName());
        assertEquals("テスト", dto.content());
    }
}
