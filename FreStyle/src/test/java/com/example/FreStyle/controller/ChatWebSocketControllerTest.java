package com.example.FreStyle.controller;

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
import org.springframework.messaging.simp.SimpMessagingTemplate;

import java.sql.Timestamp;
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

    @BeforeEach
    void setUp() {
        savedMessage = new ChatMessageDto();
        savedMessage.setId(100);
        savedMessage.setRoomId(10);
        savedMessage.setSenderId(1);
        savedMessage.setSenderName("送信者");
        savedMessage.setContent("テストメッセージ");
        savedMessage.setCreatedAt(new Timestamp(System.currentTimeMillis()));
    }

    @Test
    @DisplayName("sendMessage: 相手のunreadCountがWebSocket通知される")
    void sendMessage_sendsUnreadNotificationViaWebSocket() {
        when(sendChatMessageUseCase.execute(1, 10, "テストメッセージ"))
                .thenReturn(new SendChatMessageUseCase.Result(savedMessage, 2));

        Map<String, Object> payload = Map.of(
                "senderId", 1,
                "roomId", 10,
                "content", "テストメッセージ"
        );

        controller.sendMessage(payload);

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
                "senderId", 1,
                "roomId", 10,
                "content", "テストメッセージ"
        );

        controller.sendMessage(payload);

        verify(messagingTemplate).convertAndSend(eq("/topic/chat/10"), eq(savedMessage));
        verify(messagingTemplate, times(1)).convertAndSend(anyString(), any(Object.class));
    }

    @Test
    @DisplayName("deleteMessage: 削除通知をWebSocketで送信する")
    void deleteMessage_sendsDeleteNotificationViaWebSocket() {
        Map<String, Object> payload = Map.of(
                "messageId", 100,
                "roomId", 10
        );

        controller.deleteMessage(payload);

        verify(deleteChatMessageUseCase).execute(100);

        @SuppressWarnings("unchecked")
        ArgumentCaptor<Map<String, Object>> captor = ArgumentCaptor.forClass(Map.class);
        verify(messagingTemplate).convertAndSend(eq("/topic/chat/10"), captor.capture());
        Map<String, Object> notification = captor.getValue();
        assertEquals("delete", notification.get("type"));
        assertEquals(100, notification.get("messageId"));
    }
}
