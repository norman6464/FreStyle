package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import java.util.Optional;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.ChatMessageDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.RoomMemberRepository;
import com.example.FreStyle.service.ChatMessageService;
import com.example.FreStyle.service.UnreadCountService;

@ExtendWith(MockitoExtension.class)
class SendChatMessageUseCaseTest {

    @Mock
    private ChatMessageService chatMessageService;

    @Mock
    private UnreadCountService unreadCountService;

    @Mock
    private RoomMemberRepository roomMemberRepository;

    @InjectMocks
    private SendChatMessageUseCase useCase;

    private ChatMessageDto savedMessage;
    private User partner;

    @BeforeEach
    void setUp() {
        savedMessage = new ChatMessageDto("msg-100", 10, 1, "送信者", "テストメッセージ", 1000L);

        partner = new User();
        partner.setId(2);
        partner.setName("相手");
    }

    @Test
    @DisplayName("メッセージを保存しChatMessageDtoと相手IDを返す")
    void execute_savesMessageAndReturnsResult() {
        when(chatMessageService.addMessage(10, 1, "テストメッセージ")).thenReturn(savedMessage);
        when(roomMemberRepository.findPartnerByRoomIdAndUserId(10, 1)).thenReturn(Optional.of(partner));

        SendChatMessageUseCase.Result result = useCase.execute(1, 10, "テストメッセージ");

        assertEquals(savedMessage, result.message());
        assertEquals(2, result.partnerId());
        verify(unreadCountService).incrementUnreadCount(2, 10);
    }

    @Test
    @DisplayName("addMessage例外時にそのまま伝搬する")
    void propagatesExceptionWhenAddMessageFails() {
        when(chatMessageService.addMessage(999, 1, "テストメッセージ"))
                .thenThrow(new RuntimeException("保存失敗"));

        assertThatThrownBy(() -> useCase.execute(1, 999, "テストメッセージ"))
                .isInstanceOf(RuntimeException.class)
                .hasMessage("保存失敗");
    }

    @Test
    @DisplayName("addMessageに正しい引数が渡される")
    void passesCorrectArgumentsToAddMessage() {
        when(chatMessageService.addMessage(10, 5, "こんにちは")).thenReturn(savedMessage);
        when(roomMemberRepository.findPartnerByRoomIdAndUserId(10, 5)).thenReturn(Optional.empty());

        useCase.execute(5, 10, "こんにちは");

        verify(chatMessageService).addMessage(10, 5, "こんにちは");
    }

    @Test
    @DisplayName("相手がいない場合はpartnerIdがnullで未読数インクリメントしない")
    void execute_noPartner_returnsNullPartnerId() {
        when(chatMessageService.addMessage(10, 1, "テストメッセージ")).thenReturn(savedMessage);
        when(roomMemberRepository.findPartnerByRoomIdAndUserId(10, 1)).thenReturn(Optional.empty());

        SendChatMessageUseCase.Result result = useCase.execute(1, 10, "テストメッセージ");

        assertEquals(savedMessage, result.message());
        assertNull(result.partnerId());
        verify(unreadCountService, never()).incrementUnreadCount(anyInt(), anyInt());
    }
}
