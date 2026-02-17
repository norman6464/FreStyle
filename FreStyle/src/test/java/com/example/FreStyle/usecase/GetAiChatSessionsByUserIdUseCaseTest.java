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

import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.mapper.AiChatSessionMapper;
import com.example.FreStyle.repository.AiChatSessionRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetAiChatSessionsByUserIdUseCase")
class GetAiChatSessionsByUserIdUseCaseTest {

    @Mock
    private AiChatSessionRepository aiChatSessionRepository;

    @Mock
    private AiChatSessionMapper mapper;

    @InjectMocks
    private GetAiChatSessionsByUserIdUseCase useCase;

    @Test
    @DisplayName("ユーザーIDでセッション一覧を取得する")
    void shouldReturnSessionsByUserId() {
        AiChatSession session1 = new AiChatSession();
        session1.setId(1);
        AiChatSession session2 = new AiChatSession();
        session2.setId(2);
        when(aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of(session1, session2));

        AiChatSessionDto dto1 = new AiChatSessionDto(1, null, null, null, null, null, null, null, null);
        AiChatSessionDto dto2 = new AiChatSessionDto(2, null, null, null, null, null, null, null, null);
        when(mapper.toDto(session1)).thenReturn(dto1);
        when(mapper.toDto(session2)).thenReturn(dto2);

        List<AiChatSessionDto> result = useCase.execute(1);

        assertThat(result).hasSize(2);
        assertThat(result.get(0).id()).isEqualTo(1);
        assertThat(result.get(1).id()).isEqualTo(2);
    }

    @Test
    @DisplayName("セッションがない場合は空リストを返す")
    void shouldReturnEmptyListWhenNoSessions() {
        when(aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(999))
                .thenReturn(Collections.emptyList());

        List<AiChatSessionDto> result = useCase.execute(999);

        assertThat(result).isEmpty();
    }
}
