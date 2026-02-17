package com.example.FreStyle.usecase;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import java.util.Map;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.RoomMemberService;
import com.example.FreStyle.service.UserIdentityService;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetChatStatsUseCase テスト")
class GetChatStatsUseCaseTest {

    @Mock
    private UserIdentityService userIdentityService;

    @Mock
    private RoomMemberService roomMemberService;

    @InjectMocks
    private GetChatStatsUseCase getChatStatsUseCase;

    @Test
    @DisplayName("チャット統計情報を取得できる")
    void execute_returnsStats() {
        User user = new User();
        user.setId(1);
        user.setName("テストユーザー");
        user.setEmail("test@example.com");
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        when(roomMemberService.countChatPartners(1)).thenReturn(5L);

        Map<String, Object> result = getChatStatsUseCase.execute("sub-123");

        assertEquals(5L, result.get("chatPartnerCount"));
        assertEquals("test@example.com", result.get("email"));
        assertEquals("テストユーザー", result.get("username"));
    }

    @Test
    @DisplayName("チャットパートナーが0人の場合")
    void execute_noPartners_returnsZero() {
        User user = new User();
        user.setId(2);
        user.setName("新規ユーザー");
        user.setEmail("new@example.com");
        when(userIdentityService.findUserBySub("sub-new")).thenReturn(user);
        when(roomMemberService.countChatPartners(2)).thenReturn(0L);

        Map<String, Object> result = getChatStatsUseCase.execute("sub-new");

        assertEquals(0L, result.get("chatPartnerCount"));
    }

    @Test
    @DisplayName("返却MapにchatPartnerCount・email・usernameの3キーが含まれる")
    void execute_returnsAllExpectedKeys() {
        User user = new User();
        user.setId(3);
        user.setName("キー検証ユーザー");
        user.setEmail("keys@example.com");
        when(userIdentityService.findUserBySub("sub-keys")).thenReturn(user);
        when(roomMemberService.countChatPartners(3)).thenReturn(10L);

        Map<String, Object> result = getChatStatsUseCase.execute("sub-keys");

        assertEquals(3, result.size());
        assertTrue(result.containsKey("chatPartnerCount"));
        assertTrue(result.containsKey("email"));
        assertTrue(result.containsKey("username"));
    }

    @Test
    @DisplayName("正しいsubでUserIdentityServiceとRoomMemberServiceを呼び出す")
    void execute_callsServicesWithCorrectParams() {
        User user = new User();
        user.setId(7);
        user.setName("検証ユーザー");
        user.setEmail("verify@example.com");
        when(userIdentityService.findUserBySub("sub-verify")).thenReturn(user);
        when(roomMemberService.countChatPartners(7)).thenReturn(3L);

        getChatStatsUseCase.execute("sub-verify");

        verify(userIdentityService).findUserBySub("sub-verify");
        verify(roomMemberService).countChatPartners(7);
    }
}
