package com.example.FreStyle.controller;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.*;

import java.util.List;
import java.util.Map;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.security.oauth2.jwt.Jwt;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.dto.ChatUserDto;
import com.example.FreStyle.dto.UserDto;
import com.example.FreStyle.usecase.CreateOrGetChatRoomUseCase;
import com.example.FreStyle.usecase.GetChatHistoryUseCase;
import com.example.FreStyle.usecase.GetChatRoomsUseCase;
import com.example.FreStyle.usecase.GetChatStatsUseCase;
import com.example.FreStyle.usecase.GetChatUsersUseCase;
import com.example.FreStyle.usecase.MarkChatAsReadUseCase;

@ExtendWith(MockitoExtension.class)
@DisplayName("ChatController")
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

    @Test
    @DisplayName("users: 正常リクエスト → ユーザー一覧を返す")
    void users_validRequest_returnsUsers() {
        Jwt jwt = createMockJwt();
        when(getChatUsersUseCase.execute("cognito-sub-123", null)).thenReturn(List.of());

        ResponseEntity<?> response = chatController.users(jwt, null);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        @SuppressWarnings("unchecked")
        Map<String, List<UserDto>> body = (Map<String, List<UserDto>>) response.getBody();
        assertThat(body.get("users")).isNotNull();
    }

    @Test
    @DisplayName("users: 検索クエリ付きリクエスト")
    void users_withQuery_passesQueryToUseCase() {
        Jwt jwt = createMockJwt();
        when(getChatUsersUseCase.execute("cognito-sub-123", "テスト")).thenReturn(List.of());

        chatController.users(jwt, "テスト");

        verify(getChatUsersUseCase).execute("cognito-sub-123", "テスト");
    }

    @Test
    @DisplayName("create: 正常リクエスト → roomIdとsuccessを返す")
    void create_validRequest_returnsRoomId() {
        Jwt jwt = createMockJwt();
        when(createOrGetChatRoomUseCase.execute("cognito-sub-123", 2)).thenReturn(99);

        ResponseEntity<?> response = chatController.create(jwt, 2);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        @SuppressWarnings("unchecked")
        Map<String, Object> body = (Map<String, Object>) response.getBody();
        assertThat(body.get("roomId")).isEqualTo(99);
        assertThat(body.get("status")).isEqualTo("success");
    }

    @Test
    @DisplayName("create: 例外発生時は例外がスローされる")
    void create_useCaseThrows_propagatesException() {
        Jwt jwt = createMockJwt();
        when(createOrGetChatRoomUseCase.execute("cognito-sub-123", 2)).thenThrow(new RuntimeException("DB error"));

        assertThatThrownBy(() -> chatController.create(jwt, 2))
                .isInstanceOf(RuntimeException.class);
    }

    @Test
    @DisplayName("history: 正常リクエスト → メッセージ履歴を返す")
    void history_validRequest_returnsHistory() {
        Jwt jwt = createMockJwt();
        when(getChatHistoryUseCase.execute("cognito-sub-123", 10)).thenReturn(List.of());

        ResponseEntity<?> response = chatController.history(jwt, 10);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        @SuppressWarnings("unchecked")
        List<ChatMessageDto> body = (List<ChatMessageDto>) response.getBody();
        assertThat(body).isNotNull();
    }

    @Test
    @DisplayName("stats: 正常リクエスト → 統計情報を返す")
    void stats_validRequest_returnsStats() {
        Jwt jwt = createMockJwt();
        Map<String, Object> stats = Map.of("chatPartnerCount", 5L, "email", "test@example.com", "username", "テストユーザー");
        when(getChatStatsUseCase.execute("cognito-sub-123")).thenReturn(stats);

        ResponseEntity<?> response = chatController.stats(jwt);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        @SuppressWarnings("unchecked")
        Map<String, Object> body = (Map<String, Object>) response.getBody();
        assertThat(body.get("chatPartnerCount")).isEqualTo(5L);
        assertThat(body.get("username")).isEqualTo("テストユーザー");
    }

    @Test
    @DisplayName("markAsRead: 正常リクエスト → 200 OKとsuccessを返す")
    void markAsRead_validRequest_returnsOk() {
        Jwt jwt = createMockJwt();

        ResponseEntity<?> response = chatController.markAsRead(jwt, 10);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        @SuppressWarnings("unchecked")
        Map<String, Object> body = (Map<String, Object>) response.getBody();
        assertThat(body.get("status")).isEqualTo("success");
        verify(markChatAsReadUseCase).execute("cognito-sub-123", 10);
    }

    @Test
    @DisplayName("getChatRooms: 正常リクエスト → チャットルーム一覧を返す")
    void getChatRooms_validRequest_returnsChatRooms() {
        Jwt jwt = createMockJwt();
        when(getChatRoomsUseCase.execute("cognito-sub-123", null)).thenReturn(List.of());

        ResponseEntity<?> response = chatController.getChatRooms(jwt, null);

        assertThat(response.getStatusCode()).isEqualTo(HttpStatus.OK);
        @SuppressWarnings("unchecked")
        Map<String, Object> body = (Map<String, Object>) response.getBody();
        assertThat(body.get("chatUsers")).isNotNull();
    }
}
