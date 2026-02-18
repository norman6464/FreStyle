package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.*;

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
@DisplayName("GetAiChatMessagesByUserIdUseCase")
class GetAiChatMessagesByUserIdUseCaseTest {

    @Mock
    private AiChatMessageDynamoRepository aiChatMessageDynamoRepository;

    @InjectMocks
    private GetAiChatMessagesByUserIdUseCase useCase;

    @Test
    @DisplayName("ユーザーのメッセージ一覧を正常に取得できる")
    void returnsMessages() {
        AiChatMessageResponseDto dto1 = new AiChatMessageResponseDto("msg-1", 1, 10, "user", "質問", 1000L);
        AiChatMessageResponseDto dto2 = new AiChatMessageResponseDto("msg-2", 2, 10, "assistant", "回答", 2000L);
        when(aiChatMessageDynamoRepository.findByUserId(10)).thenReturn(List.of(dto1, dto2));

        List<AiChatMessageResponseDto> result = useCase.execute(10);

        assertThat(result).hasSize(2);
        assertThat(result.get(0).id()).isEqualTo("msg-1");
        assertThat(result.get(1).id()).isEqualTo("msg-2");
        verify(aiChatMessageDynamoRepository).findByUserId(10);
    }

    @Test
    @DisplayName("メッセージが0件の場合は空リストを返す")
    void returnsEmptyList() {
        when(aiChatMessageDynamoRepository.findByUserId(99)).thenReturn(Collections.emptyList());

        List<AiChatMessageResponseDto> result = useCase.execute(99);

        assertThat(result).isEmpty();
    }
}
