package com.example.FreStyle.service;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.entity.ChatMessage;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.ChatMessageRepository;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.sql.Timestamp;
import java.util.List;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class ChatMessageServiceTest {

    @Mock
    private ChatMessageRepository chatMessageRepository;

    @Mock
    private UserService userService;

    @InjectMocks
    private ChatMessageService chatMessageService;

    private User createUser(Integer id, String name) {
        User user = new User();
        user.setId(id);
        user.setName(name);
        return user;
    }

    private ChatRoom createRoom(Integer id) {
        ChatRoom room = new ChatRoom();
        room.setId(id);
        return room;
    }

    private ChatMessage createMessage(Integer id, ChatRoom room, User sender, String content) {
        ChatMessage msg = new ChatMessage();
        msg.setId(id);
        msg.setRoom(room);
        msg.setSender(sender);
        msg.setContent(content);
        msg.setCreatedAt(new Timestamp(System.currentTimeMillis()));
        msg.setUpdatedAt(new Timestamp(System.currentTimeMillis()));
        return msg;
    }

    @Test
    @DisplayName("getMessagesByRoom: メッセージ一覧を返す")
    void getMessagesByRoom_returnsList() {
        ChatRoom room = createRoom(1);
        User user = createUser(10, "テストユーザー");
        ChatMessage msg1 = createMessage(1, room, user, "こんにちは");
        ChatMessage msg2 = createMessage(2, room, user, "お元気ですか");
        when(chatMessageRepository.findByRoomOrderByCreatedAtAsc(room))
                .thenReturn(List.of(msg1, msg2));

        List<ChatMessageDto> result = chatMessageService.getMessagesByRoom(room, 10);

        assertEquals(2, result.size());
        assertEquals("こんにちは", result.get(0).content());
        assertEquals("お元気ですか", result.get(1).content());
    }

    @Test
    @DisplayName("addMessage: メッセージを保存してDTOを返す")
    void addMessage_savesAndReturnsDto() {
        ChatRoom room = createRoom(1);
        User sender = createUser(10, "送信者");
        when(userService.findUserById(10)).thenReturn(sender);

        ChatMessage saved = createMessage(100, room, sender, "新メッセージ");
        when(chatMessageRepository.save(any())).thenReturn(saved);

        ChatMessageDto dto = chatMessageService.addMessage(room, 10, "新メッセージ");

        assertEquals(100, dto.id());
        assertEquals(1, dto.roomId());
        assertEquals(10, dto.senderId());
        assertEquals("新メッセージ", dto.content());
    }

    @Test
    @DisplayName("updateMessage: メッセージを更新してDTOを返す")
    void updateMessage_updatesAndReturnsDto() {
        ChatRoom room = createRoom(1);
        User user = createUser(10, "テスト");
        ChatMessage msg = createMessage(1, room, user, "旧内容");
        when(chatMessageRepository.findById(1)).thenReturn(Optional.of(msg));
        when(chatMessageRepository.save(any())).thenAnswer(inv -> inv.getArgument(0));

        ChatMessageDto dto = chatMessageService.updateMessage(1, "新内容");

        assertEquals("新内容", dto.content());
    }

    @Test
    @DisplayName("updateMessage: 存在しないメッセージで例外")
    void updateMessage_throwsWhenNotFound() {
        when(chatMessageRepository.findById(999)).thenReturn(Optional.empty());

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> chatMessageService.updateMessage(999, "更新"));
        assertEquals("メッセージが見つかりません。", ex.getMessage());
    }

    @Test
    @DisplayName("deleteMessage: メッセージを削除する")
    void deleteMessage_deletesMessage() {
        when(chatMessageRepository.existsById(1)).thenReturn(true);

        chatMessageService.deleteMessage(1);

        verify(chatMessageRepository).deleteById(1);
    }

    @Test
    @DisplayName("deleteMessage: 存在しないメッセージで例外")
    void deleteMessage_throwsWhenNotFound() {
        when(chatMessageRepository.existsById(999)).thenReturn(false);

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> chatMessageService.deleteMessage(999));
        assertEquals("メッセージが見つかりません。", ex.getMessage());
    }

    @Test
    @DisplayName("addMessage: UserServiceが例外をスローした場合そのまま伝搬する")
    void addMessage_propagatesUserServiceException() {
        ChatRoom room = createRoom(1);
        when(userService.findUserById(999))
                .thenThrow(new RuntimeException("ユーザーが見つかりません"));

        assertThrows(RuntimeException.class,
                () -> chatMessageService.addMessage(room, 999, "テスト"));
    }

    @Test
    @DisplayName("addMessage: リポジトリ保存失敗時に例外が伝搬する")
    void addMessage_propagatesRepositorySaveException() {
        ChatRoom room = createRoom(1);
        User sender = createUser(10, "送信者");
        when(userService.findUserById(10)).thenReturn(sender);
        when(chatMessageRepository.save(any()))
                .thenThrow(new RuntimeException("保存失敗"));

        assertThrows(RuntimeException.class,
                () -> chatMessageService.addMessage(room, 10, "テスト"));
    }

    @Test
    @DisplayName("getMessagesByRoom: 空のリストを返す")
    void getMessagesByRoom_returnsEmptyList() {
        ChatRoom room = createRoom(1);
        when(chatMessageRepository.findByRoomOrderByCreatedAtAsc(room))
                .thenReturn(List.of());

        List<ChatMessageDto> result = chatMessageService.getMessagesByRoom(room, 10);

        assertTrue(result.isEmpty());
    }
}
