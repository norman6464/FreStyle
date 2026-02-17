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

import com.example.FreStyle.dto.ChatUserDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.service.ChatService;
import com.example.FreStyle.service.UserIdentityService;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetChatRoomsUseCase テスト")
class GetChatRoomsUseCaseTest {

    @Mock
    private UserIdentityService userIdentityService;

    @Mock
    private ChatService chatService;

    @InjectMocks
    private GetChatRoomsUseCase getChatRoomsUseCase;

    @Test
    @DisplayName("チャットルーム一覧を取得できる")
    void execute_returnsChatRooms() {
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        ChatUserDto dto = new ChatUserDto(2, "test@example.com", "テスト", 10);
        when(chatService.findChatUsers(1, null)).thenReturn(List.of(dto));

        List<ChatUserDto> result = getChatRoomsUseCase.execute("sub-123", null);

        assertEquals(1, result.size());
        assertEquals("テスト", result.get(0).getName());
    }

    @Test
    @DisplayName("検索クエリ付きでチャットルーム一覧を取得できる")
    void execute_withQuery_passesQueryToService() {
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("sub-123")).thenReturn(user);
        when(chatService.findChatUsers(1, "テスト")).thenReturn(List.of());

        getChatRoomsUseCase.execute("sub-123", "テスト");

        verify(chatService).findChatUsers(1, "テスト");
    }

    @Test
    @DisplayName("チャットルームが0件の場合は空リストを返す")
    void execute_noRooms_returnsEmptyList() {
        User user = new User();
        user.setId(5);
        when(userIdentityService.findUserBySub("sub-empty")).thenReturn(user);
        when(chatService.findChatUsers(5, null)).thenReturn(List.of());

        List<ChatUserDto> result = getChatRoomsUseCase.execute("sub-empty", null);

        assertTrue(result.isEmpty());
    }

    @Test
    @DisplayName("複数のチャットルームを正しい順序で取得できる")
    void execute_multipleRooms_returnsAll() {
        User user = new User();
        user.setId(1);
        when(userIdentityService.findUserBySub("sub-multi")).thenReturn(user);
        ChatUserDto dto1 = new ChatUserDto(2, "a@example.com", "ユーザーA", 10);
        ChatUserDto dto2 = new ChatUserDto(3, "b@example.com", "ユーザーB", 20);
        ChatUserDto dto3 = new ChatUserDto(4, "c@example.com", "ユーザーC", 30);
        when(chatService.findChatUsers(1, null)).thenReturn(List.of(dto1, dto2, dto3));

        List<ChatUserDto> result = getChatRoomsUseCase.execute("sub-multi", null);

        assertEquals(3, result.size());
        assertEquals("ユーザーA", result.get(0).getName());
        assertEquals("ユーザーB", result.get(1).getName());
        assertEquals("ユーザーC", result.get(2).getName());
    }
}
