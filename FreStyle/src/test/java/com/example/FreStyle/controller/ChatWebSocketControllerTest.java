package com.example.FreStyle.controller;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.RoomMemberRepository;
import com.example.FreStyle.service.ChatMessageService;
import com.example.FreStyle.service.ChatRoomService;
import com.example.FreStyle.service.UnreadCountService;
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
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class ChatWebSocketControllerTest {

    @Mock
    private ChatRoomService chatRoomService;

    @Mock
    private ChatMessageService chatMessageService;

    @Mock
    private SimpMessagingTemplate messagingTemplate;

    @Mock
    private UnreadCountService unreadCountService;

    @Mock
    private RoomMemberRepository roomMemberRepository;

    @InjectMocks
    private ChatWebSocketController controller;

    private ChatRoom testRoom;
    private User sender;
    private User partner;
    private ChatMessageDto savedMessage;

    @BeforeEach
    void setUp() {
        testRoom = new ChatRoom();
        testRoom.setId(10);

        sender = new User();
        sender.setId(1);
        sender.setName("送信者");

        partner = new User();
        partner.setId(2);
        partner.setName("相手");

        savedMessage = new ChatMessageDto();
        savedMessage.setId(100);
        savedMessage.setRoomId(10);
        savedMessage.setSenderId(1);
        savedMessage.setSenderName("送信者");
        savedMessage.setContent("テストメッセージ");
        savedMessage.setCreatedAt(new Timestamp(System.currentTimeMillis()));
    }

    @Test
    @DisplayName("sendMessage: 相手のunreadCountがインクリメントされる")
    void sendMessage_incrementsPartnerUnreadCount() {
        // Arrange
        when(chatRoomService.findChatRoomById(10)).thenReturn(testRoom);
        when(chatMessageService.addMessage(testRoom, 1, "テストメッセージ")).thenReturn(savedMessage);
        when(roomMemberRepository.findPartnerByRoomIdAndUserId(10, 1)).thenReturn(Optional.of(partner));

        Map<String, Object> payload = Map.of(
                "senderId", 1,
                "roomId", 10,
                "content", "テストメッセージ"
        );

        // Act
        controller.sendMessage(payload);

        // Assert
        verify(unreadCountService).incrementUnreadCount(2, 10);
    }

    @Test
    @DisplayName("sendMessage: WebSocketで/topic/unread/{partnerId}に通知が送信される")
    void sendMessage_sendsUnreadNotificationViaWebSocket() {
        // Arrange
        when(chatRoomService.findChatRoomById(10)).thenReturn(testRoom);
        when(chatMessageService.addMessage(testRoom, 1, "テストメッセージ")).thenReturn(savedMessage);
        when(roomMemberRepository.findPartnerByRoomIdAndUserId(10, 1)).thenReturn(Optional.of(partner));

        Map<String, Object> payload = Map.of(
                "senderId", 1,
                "roomId", 10,
                "content", "テストメッセージ"
        );

        // Act
        controller.sendMessage(payload);

        // Assert - チャットメッセージのブロードキャスト
        verify(messagingTemplate).convertAndSend(eq("/topic/chat/10"), eq(savedMessage));

        // Assert - 未読数通知
        ArgumentCaptor<Map> captor = ArgumentCaptor.forClass(Map.class);
        verify(messagingTemplate).convertAndSend(eq("/topic/unread/2"), captor.capture());
        Map<String, Object> notification = captor.getValue();
        assertEquals("unread_update", notification.get("type"));
        assertEquals(10, notification.get("roomId"));
        assertEquals(1, notification.get("increment"));
    }

    @Test
    @DisplayName("sendMessage: 相手が見つからない場合でも例外が発生しない")
    void sendMessage_noPartner_doesNotThrow() {
        // Arrange
        when(chatRoomService.findChatRoomById(10)).thenReturn(testRoom);
        when(chatMessageService.addMessage(testRoom, 1, "テストメッセージ")).thenReturn(savedMessage);
        when(roomMemberRepository.findPartnerByRoomIdAndUserId(10, 1)).thenReturn(Optional.empty());

        Map<String, Object> payload = Map.of(
                "senderId", 1,
                "roomId", 10,
                "content", "テストメッセージ"
        );

        // Act & Assert - 例外なし
        assertDoesNotThrow(() -> controller.sendMessage(payload));

        // 未読数インクリメントは呼ばれない
        verify(unreadCountService, never()).incrementUnreadCount(anyInt(), anyInt());
    }
}
