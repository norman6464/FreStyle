package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.*;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.repository.AiChatMessageDynamoRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("CountAiChatMessagesBySessionIdUseCase")
class CountAiChatMessagesBySessionIdUseCaseTest {

    @Mock
    private AiChatMessageDynamoRepository aiChatMessageDynamoRepository;

    @InjectMocks
    private CountAiChatMessagesBySessionIdUseCase useCase;

    @Test
    @DisplayName("セッションのメッセージ数を正常に取得できる")
    void returnsCount() {
        when(aiChatMessageDynamoRepository.countBySessionId(5)).thenReturn(10L);

        Long result = useCase.execute(5);

        assertThat(result).isEqualTo(10L);
        verify(aiChatMessageDynamoRepository).countBySessionId(5);
    }

    @Test
    @DisplayName("メッセージが0件の場合は0を返す")
    void returnsZero() {
        when(aiChatMessageDynamoRepository.countBySessionId(99)).thenReturn(0L);

        Long result = useCase.execute(99);

        assertThat(result).isEqualTo(0L);
    }

    @Test
    @DisplayName("大量メッセージ数も正しく返す")
    void returnsLargeCount() {
        when(aiChatMessageDynamoRepository.countBySessionId(1)).thenReturn(100000L);

        Long result = useCase.execute(1);

        assertThat(result).isEqualTo(100000L);
    }

    @Test
    @DisplayName("repositoryが例外をスローした場合そのまま伝搬する")
    void propagatesRepositoryException() {
        when(aiChatMessageDynamoRepository.countBySessionId(5)).thenThrow(new RuntimeException("DB接続エラー"));

        assertThatThrownBy(() -> useCase.execute(5))
                .isInstanceOf(RuntimeException.class)
                .hasMessage("DB接続エラー");
    }
}
