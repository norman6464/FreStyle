package com.example.FreStyle.usecase;

import static org.mockito.Mockito.*;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UnreadCountService;
import com.example.FreStyle.service.UserIdentityService;

@ExtendWith(MockitoExtension.class)
@DisplayName("MarkChatAsReadUseCase テスト")
class MarkChatAsReadUseCaseTest {

    @Mock
    private UserIdentityService userIdentityService;

    @Mock
    private UnreadCountService unreadCountService;

    @InjectMocks
    private MarkChatAsReadUseCase markChatAsReadUseCase;

    @Test
    @DisplayName("未読数をリセットできる")
    void execute_resetsUnreadCount() {
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);

        markChatAsReadUseCase.execute("sub-123", 10);

        verify(unreadCountService).resetUnreadCount(1, 10);
    }

    @Test
    @DisplayName("正しいパラメータでリセットされる")
    void execute_callsWithCorrectParams() {
        User user = new User();
        user.setId(5);
        when(userIdentityService.findUserBySub("sub-456")).thenReturn(user);

        markChatAsReadUseCase.execute("sub-456", 25);

        verify(unreadCountService).resetUnreadCount(5, 25);
    }
}
