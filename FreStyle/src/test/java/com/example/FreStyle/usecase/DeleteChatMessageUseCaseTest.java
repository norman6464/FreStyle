package com.example.FreStyle.usecase;

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
}
