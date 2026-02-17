package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import java.sql.Timestamp;
import java.util.Optional;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.entity.ChatRoom;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.RoomMemberRepository;
import com.example.FreStyle.service.ChatMessageService;
import com.example.FreStyle.service.ChatRoomService;
import com.example.FreStyle.service.UnreadCountService;

@ExtendWith(MockitoExtension.class)
class SendChatMessageUseCaseTest {

    @Mock
    private ChatRoomService chatRoomService;

    @Mock
    private ChatMessageService chatMessageService;

    @Mock
    private UnreadCountService unreadCountService;

    @Mock
    private RoomMemberRepository roomMemberRepository;

    @InjectMocks
    private SendChatMessageUseCase useCase;

    private ChatRoom testRoom;
    private ChatMessageDto savedMessage;
    private User partner;

    @BeforeEach
    void setUp() {
        testRoom = new ChatRoom();
        testRoom.setId(10);

        savedMessage = new ChatMessageDto(100, 10, 1, "送信者", "テストメッセージ",
                new Timestamp(System.currentTimeMillis()), null);

        partner = new User();
        partner.setId(2);
        partner.setName("相手");
    }

    @Test
    @DisplayName("メッセージを保存しChatMessageDtoと相手IDを返す")
    void execute_savesMessageAndReturnsResult() {
        when(chatRoomService.findChatRoomById(10)).thenReturn(testRoom);
        when(chatMessageService.addMessage(testRoom, 1, "テストメッセージ")).thenReturn(savedMessage);
        when(roomMemberRepository.findPartnerByRoomIdAndUserId(10, 1)).thenReturn(Optional.of(partner));

        SendChatMessageUseCase.Result result = useCase.execute(1, 10, "テストメッセージ");

        assertEquals(savedMessage, result.message());
        assertEquals(2, result.partnerId());
        verify(unreadCountService).incrementUnreadCount(2, 10);
    }

    @Test
    @DisplayName("chatRoomService.findChatRoomById例外時にそのまま伝搬する")
    void propagatesExceptionWhenRoomNotFound() {
        when(chatRoomService.findChatRoomById(999))
                .thenThrow(new RuntimeException("チャットルームが見つかりません"));

        assertThatThrownBy(() -> useCase.execute(1, 999, "テストメッセージ"))
                .isInstanceOf(RuntimeException.class)
                .hasMessage("チャットルームが見つかりません");

        verify(chatMessageService, never()).addMessage(any(), anyInt(), anyString());
    }

    @Test
    @DisplayName("addMessageに正しい引数が渡される")
    void passesCorrectArgumentsToAddMessage() {
        when(chatRoomService.findChatRoomById(10)).thenReturn(testRoom);
        when(chatMessageService.addMessage(testRoom, 5, "こんにちは")).thenReturn(savedMessage);
        when(roomMemberRepository.findPartnerByRoomIdAndUserId(10, 5)).thenReturn(Optional.empty());

        useCase.execute(5, 10, "こんにちは");

        verify(chatMessageService).addMessage(testRoom, 5, "こんにちは");
        verify(chatRoomService).findChatRoomById(10);
    }

    @Test
    @DisplayName("相手がいない場合はpartnerIdがnullで未読数インクリメントしない")
    void execute_noPartner_returnsNullPartnerId() {
        when(chatRoomService.findChatRoomById(10)).thenReturn(testRoom);
        when(chatMessageService.addMessage(testRoom, 1, "テストメッセージ")).thenReturn(savedMessage);
        when(roomMemberRepository.findPartnerByRoomIdAndUserId(10, 1)).thenReturn(Optional.empty());

        SendChatMessageUseCase.Result result = useCase.execute(1, 10, "テストメッセージ");

        assertEquals(savedMessage, result.message());
        assertNull(result.partnerId());
        verify(unreadCountService, never()).incrementUnreadCount(anyInt(), anyInt());
    }
}
