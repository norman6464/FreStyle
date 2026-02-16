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

import com.example.FreStyle.dto.UserDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.UserIdentityService;
import com.example.FreStyle.service.UserService;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetChatUsersUseCase テスト")
class GetChatUsersUseCaseTest {

    @Mock
    private UserIdentityService userIdentityService;

    @Mock
    private UserService userService;

    @InjectMocks
    private GetChatUsersUseCase getChatUsersUseCase;

    @Test
    @DisplayName("ユーザー一覧を取得できる")
    void execute_returnsUsers() {
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        UserDto dto = new UserDto(2, "test@example.com", "テスト");
        when(userService.findUsersWithRoomId(1, null)).thenReturn(List.of(dto));

        List<UserDto> result = getChatUsersUseCase.execute("sub-123", null);

        assertEquals(1, result.size());
        assertEquals("テスト", result.get(0).getName());
    }

    @Test
    @DisplayName("検索クエリ付きでユーザー一覧を取得できる")
    void execute_withQuery_passesQueryToService() {
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        when(userService.findUsersWithRoomId(1, "検索")).thenReturn(List.of());

        getChatUsersUseCase.execute("sub-123", "検索");

        verify(userService).findUsersWithRoomId(1, "検索");
    }
}
