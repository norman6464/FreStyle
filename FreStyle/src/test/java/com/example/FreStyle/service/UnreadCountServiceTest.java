package com.example.FreStyle.service;

import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.UnreadCount;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.UnreadCountRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.Collections;
import java.util.List;
import java.util.Map;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class UnreadCountServiceTest {

    @Mock
    private UnreadCountRepository unreadCountRepository;

    @Mock
    private UserService userService;

    @Mock
    private ChatRoomService chatRoomService;

    @InjectMocks
    private UnreadCountService unreadCountService;

    private User testUser;
    private ChatRoom testRoom;

    @BeforeEach
    void setUp() {
        testUser = new User();
        testUser.setId(1);
        testUser.setName("テストユーザー");

        testRoom = new ChatRoom();
        testRoom.setId(10);
    }

    @Test
    @DisplayName("incrementUnreadCount: 既存レコードあり → カウントが+1される")
    void incrementUnreadCount_existingRecord_incrementsByOne() {
        // Arrange
        UnreadCount existing = new UnreadCount();
        existing.setId(100);
        existing.setUser(testUser);
        existing.setRoom(testRoom);
        existing.setCount(3);

        when(userService.findUserById(1)).thenReturn(testUser);
        when(chatRoomService.findChatRoomById(10)).thenReturn(testRoom);
        when(unreadCountRepository.findByUserAndRoom(testUser, testRoom))
                .thenReturn(Optional.of(existing));
        when(unreadCountRepository.save(any(UnreadCount.class)))
                .thenAnswer(invocation -> invocation.getArgument(0));

        // Act
        unreadCountService.incrementUnreadCount(1, 10);

        // Assert
        assertEquals(4, existing.getCount());
        verify(unreadCountRepository).save(existing);
    }

    @Test
    @DisplayName("incrementUnreadCount: レコードなし → 新規作成してカウント1")
    void incrementUnreadCount_noRecord_createsNewWithCountOne() {
        // Arrange
        when(userService.findUserById(1)).thenReturn(testUser);
        when(chatRoomService.findChatRoomById(10)).thenReturn(testRoom);
        when(unreadCountRepository.findByUserAndRoom(testUser, testRoom))
                .thenReturn(Optional.empty());
        when(unreadCountRepository.save(any(UnreadCount.class)))
                .thenAnswer(invocation -> invocation.getArgument(0));

        // Act
        unreadCountService.incrementUnreadCount(1, 10);

        // Assert
        verify(unreadCountRepository).save(argThat(uc ->
                uc.getUser().equals(testUser) &&
                uc.getRoom().equals(testRoom) &&
                uc.getCount() == 1
        ));
    }

    @Test
    @DisplayName("resetUnreadCount: レコードあり → カウントが0にリセットされる")
    void resetUnreadCount_existingRecord_resetsToZero() {
        // Arrange
        UnreadCount existing = new UnreadCount();
        existing.setId(100);
        existing.setUser(testUser);
        existing.setRoom(testRoom);
        existing.setCount(5);

        when(userService.findUserById(1)).thenReturn(testUser);
        when(chatRoomService.findChatRoomById(10)).thenReturn(testRoom);
        when(unreadCountRepository.findByUserAndRoom(testUser, testRoom))
                .thenReturn(Optional.of(existing));

        // Act
        unreadCountService.resetUnreadCount(1, 10);

        // Assert
        assertEquals(0, existing.getCount());
        verify(unreadCountRepository).save(existing);
    }

    @Test
    @DisplayName("resetUnreadCount: レコードなし → 例外なし")
    void resetUnreadCount_noRecord_doesNothing() {
        // Arrange
        when(userService.findUserById(1)).thenReturn(testUser);
        when(chatRoomService.findChatRoomById(10)).thenReturn(testRoom);
        when(unreadCountRepository.findByUserAndRoom(testUser, testRoom))
                .thenReturn(Optional.empty());

        // Act & Assert - 例外が発生しないこと
        assertDoesNotThrow(() -> unreadCountService.resetUnreadCount(1, 10));
        verify(unreadCountRepository, never()).save(any());
    }

    @Test
    @DisplayName("getUnreadCountsByUserAndRooms: 正しいMapを返す")
    void getUnreadCountsByUserAndRooms_returnsCorrectMap() {
        // Arrange
        ChatRoom room1 = new ChatRoom();
        room1.setId(10);
        ChatRoom room2 = new ChatRoom();
        room2.setId(20);

        UnreadCount uc1 = new UnreadCount();
        uc1.setRoom(room1);
        uc1.setCount(3);
        UnreadCount uc2 = new UnreadCount();
        uc2.setRoom(room2);
        uc2.setCount(7);

        when(unreadCountRepository.findByUserIdAndRoomIds(1, List.of(10, 20)))
                .thenReturn(List.of(uc1, uc2));

        // Act
        Map<Integer, Integer> result = unreadCountService.getUnreadCountsByUserAndRooms(1, List.of(10, 20));

        // Assert
        assertEquals(2, result.size());
        assertEquals(3, result.get(10));
        assertEquals(7, result.get(20));
    }

    @Test
    @DisplayName("getUnreadCountsByUserAndRooms: 空リスト → 空Mapを返す")
    void getUnreadCountsByUserAndRooms_emptyList_returnsEmptyMap() {
        // Act
        Map<Integer, Integer> result = unreadCountService.getUnreadCountsByUserAndRooms(1, Collections.emptyList());

        // Assert
        assertTrue(result.isEmpty());
        verify(unreadCountRepository, never()).findByUserIdAndRoomIds(anyInt(), anyList());
    }
}
