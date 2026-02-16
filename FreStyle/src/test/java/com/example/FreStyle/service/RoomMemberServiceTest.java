package com.example.FreStyle.service;

import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.RoomMemberRepository;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.List;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class RoomMemberServiceTest {

    @Mock
    private RoomMemberRepository roomMemberRepository;

    @InjectMocks
    private RoomMemberService roomMemberService;

    @Test
    @DisplayName("existsRoom: ルームにユーザーが存在する場合true")
    void existsRoom_returnsTrue() {
        when(roomMemberRepository.existsByRoom_IdAndUser_Id(10, 1)).thenReturn(true);

        assertTrue(roomMemberService.existsRoom(10, 1));
        verify(roomMemberRepository).existsByRoom_IdAndUser_Id(10, 1);
    }

    @Test
    @DisplayName("existsRoom: ルームにユーザーが存在しない場合false")
    void existsRoom_returnsFalse() {
        when(roomMemberRepository.existsByRoom_IdAndUser_Id(10, 1)).thenReturn(false);

        assertFalse(roomMemberService.existsRoom(10, 1));
    }

    @Test
    @DisplayName("findRoomId: ユーザーが所属するルームIDリストを返す")
    void findRoomId_returnsRoomIds() {
        when(roomMemberRepository.findRoomIdByUserId(1)).thenReturn(List.of(10, 20, 30));

        List<Integer> result = roomMemberService.findRoomId(1);

        assertEquals(3, result.size());
        assertEquals(List.of(10, 20, 30), result);
    }

    @Test
    @DisplayName("findUsers: 会話相手のユーザーリストを返す")
    void findUsers_returnsUsers() {
        User user1 = new User();
        user1.setId(2);
        user1.setName("ユーザー2");
        User user2 = new User();
        user2.setId(3);
        user2.setName("ユーザー3");

        when(roomMemberRepository.findUsersByUserId(1)).thenReturn(List.of(user1, user2));

        List<User> result = roomMemberService.findUsers(1);

        assertEquals(2, result.size());
        assertEquals("ユーザー2", result.get(0).getName());
        assertEquals("ユーザー3", result.get(1).getName());
    }

    @Test
    @DisplayName("countChatPartners: 会話相手の数を返す")
    void countChatPartners_returnsCount() {
        when(roomMemberRepository.countDistinctPartnersByUserId(1)).thenReturn(5L);

        Long result = roomMemberService.countChatPartners(1);

        assertEquals(5L, result);
        verify(roomMemberRepository).countDistinctPartnersByUserId(1);
    }

    @Test
    @DisplayName("findRoomId: ルームが存在しない場合空リストを返す")
    void findRoomId_returnsEmptyList() {
        when(roomMemberRepository.findRoomIdByUserId(999)).thenReturn(List.of());

        List<Integer> result = roomMemberService.findRoomId(999);

        assertTrue(result.isEmpty());
    }

    @Test
    @DisplayName("findUsers: リポジトリが例外をスローした場合そのまま伝搬する")
    void findUsers_propagatesRepositoryException() {
        when(roomMemberRepository.findUsersByUserId(1))
                .thenThrow(new RuntimeException("DB接続エラー"));

        assertThrows(RuntimeException.class,
                () -> roomMemberService.findUsers(1));
    }

    @Test
    @DisplayName("countChatPartners: 会話相手が0人の場合0を返す")
    void countChatPartners_returnsZero() {
        when(roomMemberRepository.countDistinctPartnersByUserId(999)).thenReturn(0L);

        Long result = roomMemberService.countChatPartners(999);

        assertEquals(0L, result);
    }
}
