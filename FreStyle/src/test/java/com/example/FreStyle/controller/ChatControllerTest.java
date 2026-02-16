package com.example.FreStyle.controller;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.dto.ChatUserDto;
import com.example.FreStyle.dto.UserDto;
import com.example.FreStyle.usecase.CreateOrGetChatRoomUseCase;
import com.example.FreStyle.usecase.GetChatHistoryUseCase;
import com.example.FreStyle.usecase.GetChatRoomsUseCase;
import com.example.FreStyle.usecase.GetChatStatsUseCase;
import com.example.FreStyle.usecase.GetChatUsersUseCase;
import com.example.FreStyle.usecase.MarkChatAsReadUseCase;
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

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class ChatControllerTest {

    @Mock
    private GetChatUsersUseCase getChatUsersUseCase;

    @Mock
    private CreateOrGetChatRoomUseCase createOrGetChatRoomUseCase;

    @Mock
    private GetChatHistoryUseCase getChatHistoryUseCase;

    @Mock
    private GetChatStatsUseCase getChatStatsUseCase;

    @Mock
    private MarkChatAsReadUseCase markChatAsReadUseCase;

    @Mock
    private GetChatRoomsUseCase getChatRoomsUseCase;

    @InjectMocks
    private ChatController chatController;

    private Jwt createMockJwt() {
        Jwt jwt = mock(Jwt.class);
        when(jwt.getSubject()).thenReturn("cognito-sub-123");
        return jwt;
    }

    // ============================
    // users
    // ============================
    @Test
    @DisplayName("users: 正常リクエスト → ユーザー一覧を返す")
    void users_validRequest_returnsUsers() {
        Jwt jwt = createMockJwt();
        when(getChatUsersUseCase.execute("cognito-sub-123", null)).thenReturn(List.of());

        ResponseEntity<?> response = chatController.users(jwt, null);

        assertEquals(HttpStatus.OK, response.getStatusCode());
        @SuppressWarnings("unchecked")
        Map<String, List<UserDto>> body = (Map<String, List<UserDto>>) response.getBody();
        assertNotNull(body.get("users"));
    }

    @Test
    @DisplayName("users: 検索クエリ付きリクエスト")
    void users_withQuery_passesQueryToUseCase() {
        Jwt jwt = createMockJwt();
        when(getChatUsersUseCase.execute("cognito-sub-123", "テスト")).thenReturn(List.of());

        chatController.users(jwt, "テスト");

        verify(getChatUsersUseCase).execute("cognito-sub-123", "テスト");
    }

    // ============================
    // create
    // ============================
    @Test
    @DisplayName("create: 正常リクエスト → roomIdとsuccessを返す")
    void create_validRequest_returnsRoomId() {
        Jwt jwt = createMockJwt();
        when(createOrGetChatRoomUseCase.execute("cognito-sub-123", 2)).thenReturn(99);

        ResponseEntity<?> response = chatController.create(jwt, 2);

        assertEquals(HttpStatus.OK, response.getStatusCode());
        @SuppressWarnings("unchecked")
        Map<String, Object> body = (Map<String, Object>) response.getBody();
        assertEquals(99, body.get("roomId"));
        assertEquals("success", body.get("status"));
    }

    @Test
    @DisplayName("create: 例外発生時 → 500エラーを返す")
    void create_useCaseThrows_returnsError() {
        Jwt jwt = createMockJwt();
        when(createOrGetChatRoomUseCase.execute("cognito-sub-123", 2)).thenThrow(new RuntimeException("DB error"));

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
        when(getChatHistoryUseCase.execute("cognito-sub-123", 10)).thenReturn(List.of());

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
        Map<String, Object> stats = Map.of("chatPartnerCount", 5L, "email", "test@example.com", "username", "テストユーザー");
        when(getChatStatsUseCase.execute("cognito-sub-123")).thenReturn(stats);

        ResponseEntity<?> response = chatController.stats(jwt);

        assertEquals(HttpStatus.OK, response.getStatusCode());
        @SuppressWarnings("unchecked")
        Map<String, Object> body = (Map<String, Object>) response.getBody();
        assertEquals(5L, body.get("chatPartnerCount"));
        assertEquals("テストユーザー", body.get("username"));
    }

    // ============================
    // markAsRead
    // ============================
    @Test
    @DisplayName("markAsRead: 正常リクエスト → 200 OKとsuccessを返す")
    void markAsRead_validRequest_returnsOk() {
        Jwt jwt = createMockJwt();

        ResponseEntity<?> response = chatController.markAsRead(jwt, 10);

        assertEquals(HttpStatus.OK, response.getStatusCode());
        @SuppressWarnings("unchecked")
        Map<String, Object> body = (Map<String, Object>) response.getBody();
        assertEquals("success", body.get("status"));
        verify(markChatAsReadUseCase).execute("cognito-sub-123", 10);
    }

    @Test
    @DisplayName("markAsRead: JWTなし → 401 Unauthorizedを返す")
    void markAsRead_noJwt_returnsUnauthorized() {
        ResponseEntity<?> response = chatController.markAsRead(null, 10);

        assertEquals(HttpStatus.UNAUTHORIZED, response.getStatusCode());
        verify(markChatAsReadUseCase, never()).execute(anyString(), anyInt());
    }

    // ============================
    // getChatRooms
    // ============================
    @Test
    @DisplayName("getChatRooms: 正常リクエスト → チャットルーム一覧を返す")
    void getChatRooms_validRequest_returnsChatRooms() {
        Jwt jwt = createMockJwt();
        when(getChatRoomsUseCase.execute("cognito-sub-123", null)).thenReturn(List.of());

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
