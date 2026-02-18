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
        useCase.execute(10, 1000L);

        verify(chatMessageService).deleteMessage(10, 1000L);
    }

    @Test
    @DisplayName("ChatMessageServiceが例外をスローした場合そのまま伝搬する")
    void execute_propagatesServiceException() {
        doThrow(new RuntimeException("削除失敗"))
                .when(chatMessageService).deleteMessage(10, 1000L);

        assertThatThrownBy(() -> useCase.execute(10, 1000L))
                .isInstanceOf(RuntimeException.class)
                .hasMessage("削除失敗");
    }

    @Test
    @DisplayName("異なるパラメータで正しい引数が渡される")
    void execute_passesCorrectParams() {
        useCase.execute(20, 2000L);

        verify(chatMessageService).deleteMessage(20, 2000L);
    }

    @Test
    @DisplayName("deleteMessageが1回だけ呼び出される")
    void execute_callsDeleteMessageOnce() {
        useCase.execute(10, 5000L);

        verify(chatMessageService, times(1)).deleteMessage(10, 5000L);
        verifyNoMoreInteractions(chatMessageService);
    }
}
