package com.example.FreStyle.controller;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.ChatMessageService;
import com.example.FreStyle.service.ChatRoomService;
import com.example.FreStyle.service.ChatService;
import com.example.FreStyle.service.RoomMemberService;
import com.example.FreStyle.service.UnreadCountService;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import java.util.List;
import java.util.Map;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.dto.ChatUserDto;
import com.example.FreStyle.dto.UserDto;
import com.example.FreStyle.entity.ChatRoom;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class ChatControllerTest {

    @Mock
    private UserService userService;

    @Mock
    private ChatService chatService;

    @Mock
    private ChatRoomService chatRoomService;

    @Mock
    private ChatMessageService chatMessageService;

    @Mock
    private UserIdentityService userIdentityService;

    @Mock
    private RoomMemberService roomMemberService;

    @Mock
    private UnreadCountService unreadCountService;

    @InjectMocks
    private ChatController chatController;

    private User testUser;

    @BeforeEach
    void setUp() {
        testUser = new User();
        testUser.setId(1);
        testUser.setName("テストユーザー");
    }

    private Jwt createMockJwt() {
        Jwt jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn("cognito-sub-123");
        return jwt;
    }

    @Test
    @DisplayName("markAsRead: 正常リクエスト → 200 OKとsuccessを返す")
    void markAsRead_validRequest_returnsOk() {
        // Arrange
        Jwt jwt = createMockJwt();
        when(userIdentityService.findUserBySub("cognito-sub-123")).thenReturn(testUser);

        // Act
        ResponseEntity<?> response = chatController.markAsRead(jwt, 10);

        // Assert
        assertEquals(HttpStatus.OK, response.getStatusCode());
        @SuppressWarnings("unchecked")
        Map<String, Object> body = (Map<String, Object>) response.getBody();
        assertEquals("success", body.get("status"));
        verify(unreadCountService).resetUnreadCount(1, 10);
    }

    @Test
    @DisplayName("markAsRead: JWTなし → 401 Unauthorizedを返す")
    void markAsRead_noJwt_returnsUnauthorized() {
        // Act
        ResponseEntity<?> response = chatController.markAsRead(null, 10);

        // Assert
        assertEquals(HttpStatus.UNAUTHORIZED, response.getStatusCode());
        verify(unreadCountService, never()).resetUnreadCount(anyInt(), anyInt());
    }

    @Test
    @DisplayName("markAsRead: resetUnreadCountが正しいパラメータで呼ばれる")
    void markAsRead_callsResetWithCorrectParams() {
        // Arrange
        Jwt jwt = createMockJwt();
        when(userIdentityService.findUserBySub("cognito-sub-123")).thenReturn(testUser);

        // Act
        chatController.markAsRead(jwt, 25);

        // Assert
        verify(unreadCountService).resetUnreadCount(1, 25);
    }

    // ============================
    // users
    // ============================
    @Test
    @DisplayName("users: 正常リクエスト → ユーザー一覧を返す")
    void users_validRequest_returnsUsers() {
        Jwt jwt = createMockJwt();
        when(userIdentityService.findUserBySub("cognito-sub-123")).thenReturn(testUser);
        when(userService.findUsersWithRoomId(1, null)).thenReturn(List.of());

        ResponseEntity<?> response = chatController.users(jwt, null);

        assertEquals(HttpStatus.OK, response.getStatusCode());
        @SuppressWarnings("unchecked")
        Map<String, List<UserDto>> body = (Map<String, List<UserDto>>) response.getBody();
        assertNotNull(body.get("users"));
    }

    @Test
    @DisplayName("users: 検索クエリ付きリクエスト")
    void users_withQuery_passesQueryToService() {
        Jwt jwt = createMockJwt();
        when(userIdentityService.findUserBySub("cognito-sub-123")).thenReturn(testUser);
        when(userService.findUsersWithRoomId(1, "テスト")).thenReturn(List.of());

        chatController.users(jwt, "テスト");

        verify(userService).findUsersWithRoomId(1, "テスト");
    }

    // ============================
    // create
    // ============================
    @Test
    @DisplayName("create: 正常リクエスト → roomIdとsuccessを返す")
    void create_validRequest_returnsRoomId() {
        Jwt jwt = createMockJwt();
        when(userIdentityService.findUserBySub("cognito-sub-123")).thenReturn(testUser);
        when(chatService.createOrGetRoom(1, 2)).thenReturn(99);

        ResponseEntity<?> response = chatController.create(jwt, 2);

        assertEquals(HttpStatus.OK, response.getStatusCode());
        @SuppressWarnings("unchecked")
        Map<String, Object> body = (Map<String, Object>) response.getBody();
        assertEquals(99, body.get("roomId"));
        assertEquals("success", body.get("status"));
    }

    @Test
    @DisplayName("create: 例外発生時 → 500エラーを返す")
    void create_serviceThrows_returnsError() {
        Jwt jwt = createMockJwt();
        when(userIdentityService.findUserBySub("cognito-sub-123")).thenReturn(testUser);
        when(chatService.createOrGetRoom(1, 2)).thenThrow(new RuntimeException("DB error"));

        ResponseEntity<?> response = chatController.create(jwt, 2);

        assertEquals(HttpStatus.INTERNAL_SERVER_ERROR, response.getStatusCode());
    }

    // ============================
    // history
    // ============================
    @Test
    @DisplayName("history: 正常リクエスト → メッセージ履歴を返す")
    void history_validRequest_returnsHistory() {
        Jwt jwt = createMockJwt();
        when(userIdentityService.findUserBySub("cognito-sub-123")).thenReturn(testUser);
        ChatRoom chatRoom = new ChatRoom();
        chatRoom.setId(10);
        when(chatRoomService.findChatRoomById(10)).thenReturn(chatRoom);
        when(chatMessageService.getMessagesByRoom(chatRoom, 1)).thenReturn(List.of());

        ResponseEntity<?> response = chatController.history(jwt, 10);

        assertEquals(HttpStatus.OK, response.getStatusCode());
        @SuppressWarnings("unchecked")
        List<ChatMessageDto> body = (List<ChatMessageDto>) response.getBody();
        assertNotNull(body);
    }

    // ============================
    // stats
    // ============================
    @Test
    @DisplayName("stats: 正常リクエスト → 統計情報を返す")
    void stats_validRequest_returnsStats() {
        Jwt jwt = createMockJwt();
        testUser.setEmail("test@example.com");
        when(userIdentityService.findUserBySub("cognito-sub-123")).thenReturn(testUser);
        when(roomMemberService.countChatPartners(1)).thenReturn(5L);

        ResponseEntity<?> response = chatController.stats(jwt);

        assertEquals(HttpStatus.OK, response.getStatusCode());
        @SuppressWarnings("unchecked")
        Map<String, Object> body = (Map<String, Object>) response.getBody();
        assertEquals(5L, body.get("chatPartnerCount"));
        assertEquals("テストユーザー", body.get("username"));
    }

    // ============================
    // getChatRooms
    // ============================
    @Test
    @DisplayName("getChatRooms: 正常リクエスト → チャットルーム一覧を返す")
    void getChatRooms_validRequest_returnsChatRooms() {
        Jwt jwt = createMockJwt();
        when(userIdentityService.findUserBySub("cognito-sub-123")).thenReturn(testUser);
        when(chatService.findChatUsers(1, null)).thenReturn(List.of());

        ResponseEntity<?> response = chatController.getChatRooms(jwt, null);

        assertEquals(HttpStatus.OK, response.getStatusCode());
        @SuppressWarnings("unchecked")
        Map<String, Object> body = (Map<String, Object>) response.getBody();
        assertNotNull(body.get("chatUsers"));
    }

    @Test
    @DisplayName("getChatRooms: JWTなし → 401 Unauthorizedを返す")
    void getChatRooms_noJwt_returnsUnauthorized() {
        ResponseEntity<?> response = chatController.getChatRooms(null, null);

        assertEquals(HttpStatus.UNAUTHORIZED, response.getStatusCode());
    }
}
