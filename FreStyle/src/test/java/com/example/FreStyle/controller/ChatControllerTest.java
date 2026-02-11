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

import java.util.Map;

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

    private Jwt mockJwt;
    private User testUser;

    @BeforeEach
    void setUp() {
        mockJwt = mock(Jwt.class);
        when(mockJwt.getSubject()).thenReturn("cognito-sub-123");

        testUser = new User();
        testUser.setId(1);
        testUser.setName("テストユーザー");
    }

    @Test
    @DisplayName("markAsRead: 正常リクエスト → 200 OKとsuccessを返す")
    void markAsRead_validRequest_returnsOk() {
        // Arrange
        when(userIdentityService.findUserBySub("cognito-sub-123")).thenReturn(testUser);

        // Act
        ResponseEntity<?> response = chatController.markAsRead(mockJwt, 10);

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
        when(userIdentityService.findUserBySub("cognito-sub-123")).thenReturn(testUser);

        // Act
        chatController.markAsRead(mockJwt, 25);

        // Assert
        verify(unreadCountService).resetUnreadCount(1, 25);
    }
}
