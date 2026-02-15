package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.*;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.repository.AiChatMessageRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("CountAiChatMessagesBySessionIdUseCase")
class CountAiChatMessagesBySessionIdUseCaseTest {

    @Mock private AiChatMessageRepository repository;
    @InjectMocks private CountAiChatMessagesBySessionIdUseCase useCase;

    @Test
    @DisplayName("セッションのメッセージ数を正常に取得できる")
    void returnsCount() {
        when(repository.countBySessionId(5)).thenReturn(10L);

        Long result = useCase.execute(5);

        assertThat(result).isEqualTo(10L);
        verify(repository).countBySessionId(5);
    }

    @Test
    @DisplayName("メッセージが0件の場合は0を返す")
    void returnsZero() {
        when(repository.countBySessionId(99)).thenReturn(0L);

        Long result = useCase.execute(99);

        assertThat(result).isEqualTo(0L);
    }
}
