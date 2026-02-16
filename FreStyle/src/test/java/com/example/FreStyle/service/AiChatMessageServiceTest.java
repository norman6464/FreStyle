package com.example.FreStyle.service;

import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.entity.AiChatMessage;
import com.example.FreStyle.entity.AiChatMessage.Role;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.AiChatMessageRepository;
import com.example.FreStyle.repository.AiChatSessionRepository;
import com.example.FreStyle.repository.UserRepository;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.sql.Timestamp;
import java.util.List;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class AiChatMessageServiceTest {

    @Mock
    private AiChatMessageRepository aiChatMessageRepository;

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

    private AiChatMessage createMessage(Integer id, AiChatSession session, User user, Role role, String content) {
        AiChatMessage msg = new AiChatMessage();
        msg.setId(id);
        msg.setSession(session);
        msg.setUser(user);
        msg.setRole(role);
        msg.setContent(content);
        msg.setCreatedAt(new Timestamp(System.currentTimeMillis()));
        return msg;
    }

    @Test
    @DisplayName("addMessage: メッセージを保存してDTOを返す")
    void addMessage_savesAndReturnsDto() {
        AiChatSession session = createSession(1);
        User user = createUser(10);
        when(aiChatSessionRepository.findById(1)).thenReturn(Optional.of(session));
        when(userRepository.findById(10)).thenReturn(Optional.of(user));

        AiChatMessage saved = createMessage(100, session, user, Role.user, "こんにちは");
        when(aiChatMessageRepository.save(any())).thenReturn(saved);

        AiChatMessageResponseDto dto = aiChatMessageService.addMessage(1, 10, Role.user, "こんにちは");

        assertEquals(100, dto.getId());
        assertEquals(1, dto.getSessionId());
        assertEquals(10, dto.getUserId());
        assertEquals("user", dto.getRole());
        assertEquals("こんにちは", dto.getContent());

        ArgumentCaptor<AiChatMessage> captor = ArgumentCaptor.forClass(AiChatMessage.class);
        verify(aiChatMessageRepository).save(captor.capture());
        assertEquals(session, captor.getValue().getSession());
        assertEquals(user, captor.getValue().getUser());
        assertEquals(Role.user, captor.getValue().getRole());
    }

    @Test
    @DisplayName("addMessage: セッションが存在しない場合に例外")
    void addMessage_throwsWhenSessionNotFound() {
        when(aiChatSessionRepository.findById(999)).thenReturn(Optional.empty());

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> aiChatMessageService.addMessage(999, 10, Role.user, "test"));
        assertTrue(ex.getMessage().contains("セッションが見つかりません"));
    }

    @Test
    @DisplayName("addMessage: ユーザーが存在しない場合に例外")
    void addMessage_throwsWhenUserNotFound() {
        AiChatSession session = createSession(1);
        when(aiChatSessionRepository.findById(1)).thenReturn(Optional.of(session));
        when(userRepository.findById(999)).thenReturn(Optional.empty());

        RuntimeException ex = assertThrows(RuntimeException.class,
                () -> aiChatMessageService.addMessage(1, 999, Role.user, "test"));
        assertTrue(ex.getMessage().contains("ユーザーが見つかりません"));
    }

    @Test
    @DisplayName("addUserMessage: userロールでメッセージを追加")
    void addUserMessage_addsWithUserRole() {
        AiChatSession session = createSession(1);
        User user = createUser(10);
        when(aiChatSessionRepository.findById(1)).thenReturn(Optional.of(session));
        when(userRepository.findById(10)).thenReturn(Optional.of(user));

        AiChatMessage saved = createMessage(101, session, user, Role.user, "質問");
        when(aiChatMessageRepository.save(any())).thenReturn(saved);

        AiChatMessageResponseDto dto = aiChatMessageService.addUserMessage(1, 10, "質問");
        assertEquals("user", dto.getRole());
    }

    @Test
    @DisplayName("addAssistantMessage: assistantロールでメッセージを追加")
    void addAssistantMessage_addsWithAssistantRole() {
        AiChatSession session = createSession(1);
        User user = createUser(10);
        when(aiChatSessionRepository.findById(1)).thenReturn(Optional.of(session));
        when(userRepository.findById(10)).thenReturn(Optional.of(user));

        AiChatMessage saved = createMessage(102, session, user, Role.assistant, "回答");
        when(aiChatMessageRepository.save(any())).thenReturn(saved);

        AiChatMessageResponseDto dto = aiChatMessageService.addAssistantMessage(1, 10, "回答");
        assertEquals("assistant", dto.getRole());
    }

    @Test
    @DisplayName("getMessagesBySessionId: セッションのメッセージ一覧を返す")
    void getMessagesBySessionId_returnsList() {
        AiChatSession session = createSession(1);
        User user = createUser(10);
        AiChatMessage msg1 = createMessage(1, session, user, Role.user, "質問");
        AiChatMessage msg2 = createMessage(2, session, user, Role.assistant, "回答");
        when(aiChatMessageRepository.findBySessionIdOrderByCreatedAtAsc(1))
                .thenReturn(List.of(msg1, msg2));

        List<AiChatMessageResponseDto> result = aiChatMessageService.getMessagesBySessionId(1);

        assertEquals(2, result.size());
        assertEquals("質問", result.get(0).getContent());
        assertEquals("回答", result.get(1).getContent());
    }

    @Test
    @DisplayName("countMessagesBySessionId: メッセージ数を返す")
    void countMessagesBySessionId_returnsCount() {
        when(aiChatMessageRepository.countBySessionId(1)).thenReturn(5L);

        Long count = aiChatMessageService.countMessagesBySessionId(1);

        assertEquals(5L, count);
    }

    @Test
    @DisplayName("getMessagesByUserId: ユーザーの全メッセージを返す")
    void getMessagesByUserId_returnsList() {
        AiChatSession session = createSession(1);
        User user = createUser(10);
        AiChatMessage msg1 = createMessage(1, session, user, Role.user, "質問1");
        AiChatMessage msg2 = createMessage(2, session, user, Role.assistant, "回答1");
        when(aiChatMessageRepository.findByUserIdOrderByCreatedAtAsc(10))
                .thenReturn(List.of(msg1, msg2));

        List<AiChatMessageResponseDto> result = aiChatMessageService.getMessagesByUserId(10);

        assertEquals(2, result.size());
        assertEquals("質問1", result.get(0).getContent());
        assertEquals("回答1", result.get(1).getContent());
        assertEquals(10, result.get(0).getUserId());
    }

    @Test
    @DisplayName("getMessagesByUserId: メッセージがない場合は空リスト")
    void getMessagesByUserId_returnsEmptyList() {
        when(aiChatMessageRepository.findByUserIdOrderByCreatedAtAsc(999))
                .thenReturn(List.of());

        List<AiChatMessageResponseDto> result = aiChatMessageService.getMessagesByUserId(999);

        assertTrue(result.isEmpty());
    }
}
