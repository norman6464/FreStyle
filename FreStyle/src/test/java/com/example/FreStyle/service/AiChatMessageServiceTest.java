package com.example.FreStyle.service;

import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.AiChatMessageDynamoRepository;
import com.example.FreStyle.repository.AiChatSessionRepository;
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
class AiChatMessageServiceTest {

    @Mock
    private AiChatMessageDynamoRepository aiChatMessageDynamoRepository;

    @Mock
    private AiChatSessionRepository aiChatSessionRepository;

    @Mock
    private UserRepository userRepository;

    @InjectMocks
    private AiChatMessageService aiChatMessageService;

    private AiChatSession createSession(Integer id) {
        AiChatSession session = new AiChatSession();
        session.setId(id);
        return session;
    }

    private User createUser(Integer id) {
        User user = new User();
        user.setId(id);
        return user;
    }

    @Test
    @DisplayName("addMessage: メッセージを保存してDTOを返す")
    void addMessage_savesAndReturnsDto() {
        AiChatSession session = createSession(1);
        User user = createUser(10);
        when(aiChatSessionRepository.findById(1)).thenReturn(Optional.of(session));
        when(userRepository.findById(10)).thenReturn(Optional.of(user));

        AiChatMessageResponseDto expected = new AiChatMessageResponseDto("msg-1", 1, 10, "user", "こんにちは", 1000L);
        when(aiChatMessageDynamoRepository.save(1, 10, "user", "こんにちは")).thenReturn(expected);

        AiChatMessageResponseDto dto = aiChatMessageService.addMessage(1, 10, "user", "こんにちは");

        assertEquals("msg-1", dto.id());
        assertEquals(1, dto.sessionId());
        assertEquals(10, dto.userId());
        assertEquals("user", dto.role());
        assertEquals("こんにちは", dto.content());

        verify(aiChatMessageDynamoRepository).save(1, 10, "user", "こんにちは");
    }

    @Test
    @DisplayName("addMessage: セッションが存在しない場合に例外")
    void addMessage_throwsWhenSessionNotFound() {
        when(aiChatSessionRepository.findById(999)).thenReturn(Optional.empty());

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> aiChatMessageService.addMessage(999, 10, "user", "test"));
        assertTrue(ex.getMessage().contains("セッションが見つかりません"));
    }

    @Test
    @DisplayName("addMessage: ユーザーが存在しない場合に例外")
    void addMessage_throwsWhenUserNotFound() {
        AiChatSession session = createSession(1);
        when(aiChatSessionRepository.findById(1)).thenReturn(Optional.of(session));
        when(userRepository.findById(999)).thenReturn(Optional.empty());

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> aiChatMessageService.addMessage(1, 999, "user", "test"));
        assertTrue(ex.getMessage().contains("ユーザーが見つかりません"));
    }

    @Test
    @DisplayName("addUserMessage: userロールでメッセージを追加")
    void addUserMessage_addsWithUserRole() {
        AiChatSession session = createSession(1);
        User user = createUser(10);
        when(aiChatSessionRepository.findById(1)).thenReturn(Optional.of(session));
        when(userRepository.findById(10)).thenReturn(Optional.of(user));

        AiChatMessageResponseDto expected = new AiChatMessageResponseDto("msg-1", 1, 10, "user", "質問", 1000L);
        when(aiChatMessageDynamoRepository.save(1, 10, "user", "質問")).thenReturn(expected);

        AiChatMessageResponseDto dto = aiChatMessageService.addUserMessage(1, 10, "質問");
        assertEquals("user", dto.role());
    }

    @Test
    @DisplayName("addAssistantMessage: assistantロールでメッセージを追加")
    void addAssistantMessage_addsWithAssistantRole() {
        AiChatSession session = createSession(1);
        User user = createUser(10);
        when(aiChatSessionRepository.findById(1)).thenReturn(Optional.of(session));
        when(userRepository.findById(10)).thenReturn(Optional.of(user));

        AiChatMessageResponseDto expected = new AiChatMessageResponseDto("msg-1", 1, 10, "assistant", "回答", 1000L);
        when(aiChatMessageDynamoRepository.save(1, 10, "assistant", "回答")).thenReturn(expected);

        AiChatMessageResponseDto dto = aiChatMessageService.addAssistantMessage(1, 10, "回答");
        assertEquals("assistant", dto.role());
    }

    @Test
    @DisplayName("getMessagesBySessionId: セッションのメッセージ一覧を返す")
    void getMessagesBySessionId_returnsList() {
        List<AiChatMessageResponseDto> expected = List.of(
                new AiChatMessageResponseDto("msg-1", 1, 10, "user", "質問", 1000L),
                new AiChatMessageResponseDto("msg-2", 1, 10, "assistant", "回答", 2000L)
        );
        when(aiChatMessageDynamoRepository.findBySessionId(1)).thenReturn(expected);

        List<AiChatMessageResponseDto> result = aiChatMessageService.getMessagesBySessionId(1);

        assertEquals(2, result.size());
        assertEquals("質問", result.get(0).content());
        assertEquals("回答", result.get(1).content());
    }

    @Test
    @DisplayName("countMessagesBySessionId: メッセージ数を返す")
    void countMessagesBySessionId_returnsCount() {
        when(aiChatMessageDynamoRepository.countBySessionId(1)).thenReturn(5L);

        Long count = aiChatMessageService.countMessagesBySessionId(1);

        assertEquals(5L, count);
    }

    @Test
    @DisplayName("getMessagesByUserId: ユーザーの全メッセージを返す")
    void getMessagesByUserId_returnsList() {
        List<AiChatMessageResponseDto> expected = List.of(
                new AiChatMessageResponseDto("msg-1", 1, 10, "user", "質問1", 1000L),
                new AiChatMessageResponseDto("msg-2", 1, 10, "assistant", "回答1", 2000L)
        );
        when(aiChatMessageDynamoRepository.findByUserId(10)).thenReturn(expected);

        List<AiChatMessageResponseDto> result = aiChatMessageService.getMessagesByUserId(10);

        assertEquals(2, result.size());
        assertEquals("質問1", result.get(0).content());
        assertEquals("回答1", result.get(1).content());
        assertEquals(10, result.get(0).userId());
    }

    @Test
    @DisplayName("getMessagesByUserId: メッセージがない場合は空リスト")
    void getMessagesByUserId_returnsEmptyList() {
        when(aiChatMessageDynamoRepository.findByUserId(999)).thenReturn(List.of());

        List<AiChatMessageResponseDto> result = aiChatMessageService.getMessagesByUserId(999);

        assertTrue(result.isEmpty());
    }
}
