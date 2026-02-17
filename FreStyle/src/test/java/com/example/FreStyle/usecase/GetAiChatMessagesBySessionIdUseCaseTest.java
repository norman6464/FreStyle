package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
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
import com.example.FreStyle.entity.AiChatMessage;
import com.example.FreStyle.mapper.AiChatMessageMapper;
import com.example.FreStyle.repository.AiChatMessageRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetAiChatMessagesBySessionIdUseCase")
class GetAiChatMessagesBySessionIdUseCaseTest {

    @Mock
    private AiChatMessageRepository aiChatMessageRepository;

    @Mock
    private AiChatMessageMapper mapper;

    @InjectMocks
    private GetAiChatMessagesBySessionIdUseCase useCase;

    @Test
    @DisplayName("セッションIDでメッセージ一覧を取得する")
    void shouldReturnMessagesBySessionId() {
        AiChatMessage msg1 = new AiChatMessage();
        msg1.setId(1);
        AiChatMessage msg2 = new AiChatMessage();
        msg2.setId(2);
        when(aiChatMessageRepository.findBySessionIdOrderByCreatedAtAsc(1))
                .thenReturn(List.of(msg1, msg2));

        AiChatMessageResponseDto dto1 = new AiChatMessageResponseDto(1, null, null, null, null, null);
        AiChatMessageResponseDto dto2 = new AiChatMessageResponseDto(2, null, null, null, null, null);
        when(mapper.toDto(msg1)).thenReturn(dto1);
        when(mapper.toDto(msg2)).thenReturn(dto2);

        List<AiChatMessageResponseDto> result = useCase.execute(1);

        assertThat(result).hasSize(2);
        assertThat(result.get(0).id()).isEqualTo(1);
        assertThat(result.get(1).id()).isEqualTo(2);
    }

    @Test
    @DisplayName("メッセージがない場合は空リストを返す")
    void shouldReturnEmptyListWhenNoMessages() {
        when(aiChatMessageRepository.findBySessionIdOrderByCreatedAtAsc(999))
                .thenReturn(Collections.emptyList());

        List<AiChatMessageResponseDto> result = useCase.execute(999);

        assertThat(result).isEmpty();
    }
}
