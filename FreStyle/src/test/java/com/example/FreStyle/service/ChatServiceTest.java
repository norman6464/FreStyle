package com.example.FreStyle.service;

import com.example.FreStyle.dto.ChatUserDto;
import com.example.FreStyle.entity.ChatMessage;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.ChatMessageRepository;
import com.example.FreStyle.repository.ChatRoomRepository;
import com.example.FreStyle.repository.RoomMemberRepository;
import com.example.FreStyle.repository.UserRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.sql.Timestamp;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

@ExtendWith(MockitoExtension.class)
class ChatServiceTest {

    @Mock
    private ChatRoomRepository chatRoomRepository;

    @Mock
    private RoomMemberRepository roomMemberRepository;

    @Mock
    private UserRepository userRepository;

    @Mock
    private ChatMessageRepository chatMessageRepository;

    @Mock
    private UnreadCountService unreadCountService;

    @InjectMocks
    private ChatService chatService;

    private User myUser;
    private User partnerUser;
    private ChatRoom room;
    private ChatMessage latestMessage;

    @BeforeEach
    void setUp() {
        myUser = new User();
        myUser.setId(1);
        myUser.setName("自分");
        myUser.setEmail("me@test.com");

        partnerUser = new User();
        partnerUser.setId(2);
        partnerUser.setName("相手");
        partnerUser.setEmail("partner@test.com");

        room = new ChatRoom();
        room.setId(10);

        latestMessage = new ChatMessage();
        latestMessage.setId(100);
        latestMessage.setRoom(room);
        latestMessage.setSender(partnerUser);
        latestMessage.setContent("こんにちは");
        latestMessage.setCreatedAt(new Timestamp(System.currentTimeMillis()));
    }

    @Test
    @DisplayName("findChatUsers: 未読数が正しくDTOに設定される")
    void findChatUsers_setsCorrectUnreadCount() {
        // Arrange
        // partnerId=2, roomId=10 のペアを返す
        List<Object[]> partnerDataList = new ArrayList<>();
        partnerDataList.add(new Object[]{2, 10});
        when(roomMemberRepository.findPartnerUserIdAndRoomIdByUserId(1))
                .thenReturn(partnerDataList);
        when(userRepository.findAllById(List.of(2)))
                .thenReturn(List.of(partnerUser));
        when(chatMessageRepository.findLatestMessagesByRoomIds(List.of(10)))
                .thenReturn(List.of(latestMessage));

        // 未読数: room 10 に3件の未読
        when(unreadCountService.getUnreadCountsByUserAndRooms(1, List.of(10)))
                .thenReturn(Map.of(10, 3));

        // Act
        List<ChatUserDto> result = chatService.findChatUsers(1, null);

        // Assert
        assertEquals(1, result.size());
        ChatUserDto dto = result.get(0);
        assertEquals(3, dto.getUnreadCount());
        assertEquals("相手", dto.getName());
        assertEquals(10, dto.getRoomId());
    }

    @Test
    @DisplayName("findChatUsers: 未読レコードなしのルームはカウント0")
    void findChatUsers_noUnreadRecord_defaultsToZero() {
        // Arrange
        Object[] partnerData = new Object[]{2, 10};
        when(roomMemberRepository.findPartnerUserIdAndRoomIdByUserId(1))
                .thenReturn(List.of(partnerData));
        when(userRepository.findAllById(List.of(2)))
                .thenReturn(List.of(partnerUser));
        when(chatMessageRepository.findLatestMessagesByRoomIds(List.of(10)))
                .thenReturn(List.of(latestMessage));

        // 未読数: 空Map（レコードなし）
        when(unreadCountService.getUnreadCountsByUserAndRooms(1, List.of(10)))
                .thenReturn(Map.of());

        // Act
        List<ChatUserDto> result = chatService.findChatUsers(1, null);

        // Assert
        assertEquals(1, result.size());
        assertEquals(0, result.get(0).getUnreadCount());
    }
}
