package com.example.FreStyle.controller;

import com.example.FreStyle.config.WebSocketAuthHandshakeInterceptor;
import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.usecase.DeleteChatMessageUseCase;
import com.example.FreStyle.usecase.SendChatMessageUseCase;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.messaging.simp.SimpMessageHeaderAccessor;
import org.springframework.messaging.simp.SimpMessagingTemplate;

import java.util.HashMap;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class ChatWebSocketControllerTest {

    @Mock
    private SendChatMessageUseCase sendChatMessageUseCase;

    @Mock
    private DeleteChatMessageUseCase deleteChatMessageUseCase;

    @Mock
    private SimpMessagingTemplate messagingTemplate;

    @InjectMocks
    private ChatWebSocketController controller;

    private ChatMessageDto savedMessage;
    private SimpMessageHeaderAccessor headerAccessor;

    @BeforeEach
    void setUp() {
        savedMessage = new ChatMessageDto("msg-100", 10, 1, "送信者", "テストメッセージ", 1000L);
        headerAccessor = SimpMessageHeaderAccessor.create();
        Map<String, Object> sessionAttributes = new HashMap<>();
        sessionAttributes.put(WebSocketAuthHandshakeInterceptor.AUTHENTICATED_USER_ID, 1);
        headerAccessor.setSessionAttributes(sessionAttributes);
    }

    @Test
    @DisplayName("sendMessage: 相手のunreadCountがWebSocket通知される")
    void sendMessage_sendsUnreadNotificationViaWebSocket() {
        when(sendChatMessageUseCase.execute(1, 10, "テストメッセージ"))
                .thenReturn(new SendChatMessageUseCase.Result(savedMessage, 2));

        Map<String, Object> payload = Map.of(
                "roomId", 10,
                "content", "テストメッセージ"
        );

        controller.sendMessage(payload, headerAccessor);

        verify(messagingTemplate).convertAndSend(eq("/topic/chat/10"), eq(savedMessage));

        @SuppressWarnings("unchecked")
        ArgumentCaptor<Map<String, Object>> captor = ArgumentCaptor.forClass(Map.class);
        verify(messagingTemplate).convertAndSend(eq("/topic/unread/2"), captor.capture());
        Map<String, Object> notification = captor.getValue();
        assertEquals("unread_update", notification.get("type"));
        assertEquals(10, notification.get("roomId"));
        assertEquals(1, notification.get("increment"));
    }

    @Test
    @DisplayName("sendMessage: 相手が見つからない場合は未読通知を送信しない")
    void sendMessage_noPartner_doesNotSendUnreadNotification() {
        when(sendChatMessageUseCase.execute(1, 10, "テストメッセージ"))
                .thenReturn(new SendChatMessageUseCase.Result(savedMessage, null));

        Map<String, Object> payload = Map.of(
                "roomId", 10,
                "content", "テストメッセージ"
        );

        controller.sendMessage(payload, headerAccessor);

        verify(messagingTemplate).convertAndSend(eq("/topic/chat/10"), eq(savedMessage));
        verify(messagingTemplate, times(1)).convertAndSend(anyString(), any(Object.class));
    }

    @Test
    @DisplayName("sendMessage: 未認証ユーザーの場合はメッセージ送信しない")
    void sendMessage_unauthenticated_doesNotSendMessage() {
        SimpMessageHeaderAccessor unauthHeader = SimpMessageHeaderAccessor.create();
        unauthHeader.setSessionAttributes(new HashMap<>());

        Map<String, Object> payload = Map.of(
                "roomId", 10,
                "content", "テストメッセージ"
        );

        controller.sendMessage(payload, unauthHeader);

        verify(sendChatMessageUseCase, never()).execute(anyInt(), anyInt(), anyString());
        verify(messagingTemplate, never()).convertAndSend(anyString(), any(Object.class));
    }

    @Test
    @DisplayName("deleteMessage: 削除通知をWebSocketで送信する")
    void deleteMessage_sendsDeleteNotificationViaWebSocket() {
        Map<String, Object> payload = Map.of(
                "roomId", 10,
                "createdAt", 1000L
        );

        controller.deleteMessage(payload);

        verify(deleteChatMessageUseCase).execute(10, 1000L);

        @SuppressWarnings("unchecked")
        ArgumentCaptor<Map<String, Object>> captor = ArgumentCaptor.forClass(Map.class);
        verify(messagingTemplate).convertAndSend(eq("/topic/chat/10"), captor.capture());
        Map<String, Object> notification = captor.getValue();
        assertEquals("delete", notification.get("type"));
        assertEquals(10, notification.get("roomId"));
        assertEquals(1000L, notification.get("createdAt"));
    }

    @Test
    @DisplayName("sendMessage: UseCase例外時にWebSocket通知が送信されない")
    void sendMessage_exceptionDoesNotSendNotification() {
        when(sendChatMessageUseCase.execute(1, 10, "テストメッセージ"))
                .thenThrow(new RuntimeException("DB接続エラー"));

        Map<String, Object> payload = Map.of(
                "roomId", 10,
                "content", "テストメッセージ"
        );

        assertDoesNotThrow(() -> controller.sendMessage(payload, headerAccessor));
        verify(messagingTemplate, never()).convertAndSend(anyString(), any(Object.class));
    }

    @Test
    @DisplayName("deleteMessage: UseCase例外時にWebSocket通知が送信されない")
    void deleteMessage_exceptionDoesNotSendNotification() {
        doThrow(new RuntimeException("削除エラー"))
                .when(deleteChatMessageUseCase).execute(10, 1000L);

        Map<String, Object> payload = Map.of(
                "roomId", 10,
                "createdAt", 1000L
        );

        assertDoesNotThrow(() -> controller.deleteMessage(payload));
        verify(messagingTemplate, never()).convertAndSend(anyString(), any(Object.class));
    }
}
