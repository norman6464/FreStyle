package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.*;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.service.ChatMessageService;

@ExtendWith(MockitoExtension.class)
class DeleteChatMessageUseCaseTest {

    @Mock
    private ChatMessageService chatMessageService;

    @InjectMocks
    private DeleteChatMessageUseCase useCase;

    @Test
    @DisplayName("メッセージを削除する")
    void execute_deletesMessage() {
        useCase.execute(100);

        verify(chatMessageService).deleteMessage(100);
    }

    @Test
    @DisplayName("ChatMessageServiceが例外をスローした場合そのまま伝搬する")
    void execute_propagatesServiceException() {
        doThrow(new RuntimeException("削除失敗"))
                .when(chatMessageService).deleteMessage(100);

        assertThatThrownBy(() -> useCase.execute(100))
                .isInstanceOf(RuntimeException.class)
                .hasMessage("削除失敗");
    }

    @Test
    @DisplayName("異なるメッセージIDで正しいパラメータが渡される")
    void execute_passesCorrectMessageId() {
        useCase.execute(999);

        verify(chatMessageService).deleteMessage(999);
        verify(chatMessageService, never()).deleteMessage(100);
    }

    @Test
    @DisplayName("deleteMessageが1回だけ呼び出される")
    void execute_callsDeleteMessageOnce() {
        useCase.execute(50);

        verify(chatMessageService, times(1)).deleteMessage(50);
        verifyNoMoreInteractions(chatMessageService);
    }
}
