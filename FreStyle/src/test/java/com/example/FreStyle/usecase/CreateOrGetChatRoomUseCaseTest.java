package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.ChatService;
import com.example.FreStyle.service.UserIdentityService;

@ExtendWith(MockitoExtension.class)
@DisplayName("CreateOrGetChatRoomUseCase テスト")
class CreateOrGetChatRoomUseCaseTest {

    @Mock
    private UserIdentityService userIdentityService;

    @Mock
    private ChatService chatService;

    @InjectMocks
    private CreateOrGetChatRoomUseCase createOrGetChatRoomUseCase;

    @Test
    @DisplayName("チャットルームを作成または取得できる")
    void execute_returnsRoomId() {
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        when(chatService.createOrGetRoom(1, 2)).thenReturn(99);

        Integer roomId = createOrGetChatRoomUseCase.execute("sub-123", 2);

        assertEquals(99, roomId);
    }

    @Test
    @DisplayName("正しいパラメータでサービスを呼び出す")
    void execute_callsServiceWithCorrectParams() {
        User user = new User();
        user.setId(5);
        when(userIdentityService.findUserBySub("sub-456")).thenReturn(user);
        when(chatService.createOrGetRoom(5, 10)).thenReturn(42);

        createOrGetChatRoomUseCase.execute("sub-456", 10);

        verify(chatService).createOrGetRoom(5, 10);
    }

    @Test
    @DisplayName("UserIdentityServiceが例外をスローした場合そのまま伝搬する")
    void execute_propagatesUserIdentityException() {
        when(userIdentityService.findUserBySub("unknown-sub"))
                .thenThrow(new RuntimeException("ユーザーが見つかりません"));

        assertThatThrownBy(() -> createOrGetChatRoomUseCase.execute("unknown-sub", 2))
                .isInstanceOf(RuntimeException.class)
                .hasMessage("ユーザーが見つかりません");
    }

    @Test
    @DisplayName("ChatServiceが例外をスローした場合そのまま伝搬する")
    void execute_propagatesChatServiceException() {
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        when(chatService.createOrGetRoom(1, 2))
                .thenThrow(new RuntimeException("ルーム作成失敗"));

        assertThatThrownBy(() -> createOrGetChatRoomUseCase.execute("sub-123", 2))
                .isInstanceOf(RuntimeException.class)
                .hasMessage("ルーム作成失敗");
    }
}
