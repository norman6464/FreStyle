package com.example.FreStyle.usecase;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.ChatMessageService;
import com.example.FreStyle.service.UserIdentityService;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetChatHistoryUseCase テスト")
class GetChatHistoryUseCaseTest {

    @Mock
    private UserIdentityService userIdentityService;

    @Mock
    private ChatMessageService chatMessageService;

    @InjectMocks
    private GetChatHistoryUseCase getChatHistoryUseCase;

    @Test
    @DisplayName("チャット履歴を取得できる")
    void execute_returnsHistory() {
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        ChatMessageDto msg = new ChatMessageDto(null, null, null, null, "テストメッセージ", null);
        when(chatMessageService.getMessagesByRoom(10, 1)).thenReturn(List.of(msg));

        List<ChatMessageDto> result = getChatHistoryUseCase.execute("sub-123", 10);

        assertEquals(1, result.size());
        assertEquals("テストメッセージ", result.get(0).content());
    }

    @Test
    @DisplayName("正しいルームIDとユーザーIDでサービスを呼び出す")
    void execute_callsServicesWithCorrectParams() {
        User user = new User();
        user.setId(3);
        when(userIdentityService.findUserBySub("sub-789")).thenReturn(user);
        when(chatMessageService.getMessagesByRoom(20, 3)).thenReturn(List.of());

        getChatHistoryUseCase.execute("sub-789", 20);

        verify(chatMessageService).getMessagesByRoom(20, 3);
    }
}
