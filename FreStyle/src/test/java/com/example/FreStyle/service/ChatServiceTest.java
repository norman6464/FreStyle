package com.example.FreStyle.service;

import com.example.FreStyle.dto.ChatUserDto;
import com.example.FreStyle.dto.PartnerRoomProjection;
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
import static org.mockito.ArgumentMatchers.*;
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

    private PartnerRoomProjection createProjection(Integer userId, Integer roomId) {
        PartnerRoomProjection projection = mock(PartnerRoomProjection.class);
        when(projection.getUserId()).thenReturn(userId);
        when(projection.getRoomId()).thenReturn(roomId);
        return projection;
    }

    // ============================
    // createOrGetRoom
    // ============================
    @Test
    @DisplayName("createOrGetRoom: 既存ルームが見つかればそのroomIdを返す")
    void createOrGetRoom_existingRoom_returnsRoomId() {
        when(chatRoomRepository.findRoomIdByUserIds(1, 2)).thenReturn(10);

        Integer result = chatService.createOrGetRoom(1, 2);

        assertEquals(10, result);
        verify(chatRoomRepository, never()).save(any());
        verify(roomMemberRepository, never()).saveAll(anyList());
    }

    @Test
    @DisplayName("createOrGetRoom: 新規ルームを作成しメンバーを登録する")
    void createOrGetRoom_newRoom_createsRoomAndMembers() {
        when(chatRoomRepository.findRoomIdByUserIds(1, 2)).thenReturn(null);
        when(chatRoomRepository.save(any(ChatRoom.class))).thenAnswer(invocation -> {
            ChatRoom arg = invocation.getArgument(0);
            arg.setId(99);
            return arg;
        });
        when(userRepository.findById(1)).thenReturn(Optional.of(myUser));
        when(userRepository.findById(2)).thenReturn(Optional.of(partnerUser));

        Integer result = chatService.createOrGetRoom(1, 2);

        assertEquals(99, result);
        verify(chatRoomRepository).save(any(ChatRoom.class));
        verify(roomMemberRepository).saveAll(anyList());
    }

    @Test
    @DisplayName("createOrGetRoom: 自分のユーザーが見つからない場合はIllegalStateException")
    void createOrGetRoom_myUserNotFound_throwsException() {
        when(chatRoomRepository.findRoomIdByUserIds(1, 2)).thenReturn(null);
        when(chatRoomRepository.save(any(ChatRoom.class))).thenReturn(new ChatRoom());
        when(userRepository.findById(1)).thenReturn(Optional.empty());

        IllegalStateException ex = assertThrows(IllegalStateException.class,
                () -> chatService.createOrGetRoom(1, 2));
        assertEquals("ユーザーが存在しません。", ex.getMessage());
    }

    @Test
    @DisplayName("createOrGetRoom: 相手ユーザーが見つからない場合はIllegalStateException")
    void createOrGetRoom_targetUserNotFound_throwsException() {
        when(chatRoomRepository.findRoomIdByUserIds(1, 2)).thenReturn(null);
        when(chatRoomRepository.save(any(ChatRoom.class))).thenReturn(new ChatRoom());
        when(userRepository.findById(1)).thenReturn(Optional.of(myUser));
        when(userRepository.findById(2)).thenReturn(Optional.empty());

        IllegalStateException ex = assertThrows(IllegalStateException.class,
                () -> chatService.createOrGetRoom(1, 2));
        assertEquals("相手ユーザーが存在しません。", ex.getMessage());
    }

    // ============================
    // findChatUsers
    // ============================
    @Test
    @DisplayName("findChatUsers: パートナーが存在しない場合は空リストを返す")
    void findChatUsers_noPartners_returnsEmptyList() {
        when(roomMemberRepository.findPartnerUserIdAndRoomIdByUserId(1))
                .thenReturn(new ArrayList<>());

        List<ChatUserDto> result = chatService.findChatUsers(1, null);

        assertTrue(result.isEmpty());
    }

    @Test
    @DisplayName("findChatUsers: 検索クエリで名前フィルタリングできる")
    void findChatUsers_withQuery_filtersUsersByName() {
        User partner2 = new User();
        partner2.setId(3);
        partner2.setName("フィルタ対象外");
        partner2.setEmail("other@test.com");

        List<PartnerRoomProjection> partnerDataList = new ArrayList<>();
        partnerDataList.add(createProjection(2, 10));
        partnerDataList.add(createProjection(3, 20));
        when(roomMemberRepository.findPartnerUserIdAndRoomIdByUserId(1))
                .thenReturn(partnerDataList);
        when(userRepository.findAllById(List.of(2, 3)))
                .thenReturn(List.of(partnerUser, partner2));
        when(chatMessageRepository.findLatestMessagesByRoomIds(List.of(10, 20)))
                .thenReturn(List.of());
        when(unreadCountService.getUnreadCountsByUserAndRooms(eq(1), anyList()))
                .thenReturn(Map.of());

        List<ChatUserDto> result = chatService.findChatUsers(1, "相手");

        assertEquals(1, result.size());
        assertEquals("相手", result.get(0).getName());
    }

    @Test
    @DisplayName("findChatUsers: 最終メッセージ日時で降順ソートされる")
    void findChatUsers_sortedByLastMessageDesc() {
        User partner2 = new User();
        partner2.setId(3);
        partner2.setName("ユーザー3");
        partner2.setEmail("user3@test.com");

        ChatRoom room2 = new ChatRoom();
        room2.setId(20);

        ChatMessage olderMessage = new ChatMessage();
        olderMessage.setRoom(room);
        olderMessage.setSender(partnerUser);
        olderMessage.setContent("古いメッセージ");
        olderMessage.setCreatedAt(new Timestamp(1000));

        ChatMessage newerMessage = new ChatMessage();
        newerMessage.setRoom(room2);
        newerMessage.setSender(partner2);
        newerMessage.setContent("新しいメッセージ");
        newerMessage.setCreatedAt(new Timestamp(2000));

        List<PartnerRoomProjection> partnerDataList = new ArrayList<>();
        partnerDataList.add(createProjection(2, 10));
        partnerDataList.add(createProjection(3, 20));
        when(roomMemberRepository.findPartnerUserIdAndRoomIdByUserId(1))
                .thenReturn(partnerDataList);
        when(userRepository.findAllById(List.of(2, 3)))
                .thenReturn(List.of(partnerUser, partner2));
        when(chatMessageRepository.findLatestMessagesByRoomIds(List.of(10, 20)))
                .thenReturn(List.of(olderMessage, newerMessage));
        when(unreadCountService.getUnreadCountsByUserAndRooms(eq(1), anyList()))
                .thenReturn(Map.of());

        List<ChatUserDto> result = chatService.findChatUsers(1, null);

        assertEquals(2, result.size());
        assertEquals("ユーザー3", result.get(0).getName());
        assertEquals("相手", result.get(1).getName());
    }

    @Test
    @DisplayName("findChatUsers: 未読数が正しくDTOに設定される")
    void findChatUsers_setsCorrectUnreadCount() {
        // Arrange
        // partnerId=2, roomId=10 のペアを返す
        List<PartnerRoomProjection> partnerDataList = new ArrayList<>();
        partnerDataList.add(createProjection(2, 10));
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
        List<PartnerRoomProjection> partnerDataList = new ArrayList<>();
        partnerDataList.add(createProjection(2, 10));
        when(roomMemberRepository.findPartnerUserIdAndRoomIdByUserId(1))
                .thenReturn(partnerDataList);
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
