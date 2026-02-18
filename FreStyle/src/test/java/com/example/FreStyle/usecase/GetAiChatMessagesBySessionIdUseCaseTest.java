package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.Collections;
import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.AiChatMessageResponseDto;
import com.example.FreStyle.repository.AiChatMessageDynamoRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetAiChatMessagesBySessionIdUseCase")
class GetAiChatMessagesBySessionIdUseCaseTest {

    @Mock
    private AiChatMessageDynamoRepository aiChatMessageDynamoRepository;

    @InjectMocks
    private GetAiChatMessagesBySessionIdUseCase useCase;

    @Test
    @DisplayName("セッションIDでメッセージ一覧を取得する")
    void shouldReturnMessagesBySessionId() {
        AiChatMessageResponseDto dto1 = new AiChatMessageResponseDto("msg-1", 1, 10, "user", "質問", 1000L);
        AiChatMessageResponseDto dto2 = new AiChatMessageResponseDto("msg-2", 1, 10, "assistant", "回答", 2000L);
        when(aiChatMessageDynamoRepository.findBySessionId(1)).thenReturn(List.of(dto1, dto2));

        List<AiChatMessageResponseDto> result = useCase.execute(1);

        assertThat(result).hasSize(2);
        assertThat(result.get(0).id()).isEqualTo("msg-1");
        assertThat(result.get(1).id()).isEqualTo("msg-2");
    }

    @Test
    @DisplayName("メッセージがない場合は空リストを返す")
    void shouldReturnEmptyListWhenNoMessages() {
        when(aiChatMessageDynamoRepository.findBySessionId(999)).thenReturn(Collections.emptyList());

        List<AiChatMessageResponseDto> result = useCase.execute(999);

        assertThat(result).isEmpty();
    }

    @Test
    @DisplayName("repositoryが正しく呼び出される")
    void verifiesRepositoryCalls() {
        when(aiChatMessageDynamoRepository.findBySessionId(5))
                .thenReturn(List.of(new AiChatMessageResponseDto("msg-10", 5, 1, "user", "hello", 1000L)));

        useCase.execute(5);

        verify(aiChatMessageDynamoRepository).findBySessionId(5);
    }

    @Test
    @DisplayName("単一メッセージの場合でも正しく返す")
    void shouldReturnSingleMessage() {
        when(aiChatMessageDynamoRepository.findBySessionId(3))
                .thenReturn(List.of(new AiChatMessageResponseDto("msg-42", 3, 1, "user", "hello", 1000L)));

        List<AiChatMessageResponseDto> result = useCase.execute(3);

        assertThat(result).hasSize(1);
        assertThat(result.getFirst().id()).isEqualTo("msg-42");
        assertThat(result.getFirst().content()).isEqualTo("hello");
    }
}
