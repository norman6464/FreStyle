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
import com.example.FreStyle.entity.AiChatMessage;
import com.example.FreStyle.mapper.AiChatMessageMapper;
import com.example.FreStyle.repository.AiChatMessageRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetAiChatMessagesByUserIdUseCase")
class GetAiChatMessagesByUserIdUseCaseTest {

    @Mock private AiChatMessageRepository repository;
    @Mock private AiChatMessageMapper mapper;
    @InjectMocks private GetAiChatMessagesByUserIdUseCase useCase;

    @Test
    @DisplayName("ユーザーのメッセージ一覧を正常に取得できる")
    void returnsMessages() {
        AiChatMessage msg1 = new AiChatMessage();
        msg1.setId(1);
        AiChatMessage msg2 = new AiChatMessage();
        msg2.setId(2);
        AiChatMessageResponseDto dto1 = new AiChatMessageResponseDto(1, null, null, null, null, null);
        AiChatMessageResponseDto dto2 = new AiChatMessageResponseDto(2, null, null, null, null, null);

        when(repository.findByUserIdOrderByCreatedAtAsc(10)).thenReturn(List.of(msg1, msg2));
        when(mapper.toDto(msg1)).thenReturn(dto1);
        when(mapper.toDto(msg2)).thenReturn(dto2);

        List<AiChatMessageResponseDto> result = useCase.execute(10);

        assertThat(result).hasSize(2);
        assertThat(result.get(0).id()).isEqualTo(1);
        assertThat(result.get(1).id()).isEqualTo(2);
        verify(repository).findByUserIdOrderByCreatedAtAsc(10);
    }

    @Test
    @DisplayName("メッセージが0件の場合は空リストを返す")
    void returnsEmptyList() {
        when(repository.findByUserIdOrderByCreatedAtAsc(99)).thenReturn(Collections.emptyList());

        List<AiChatMessageResponseDto> result = useCase.execute(99);

        assertThat(result).isEmpty();
    }
}
