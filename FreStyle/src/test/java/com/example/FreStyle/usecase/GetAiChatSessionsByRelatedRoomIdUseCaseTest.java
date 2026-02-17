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

import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.mapper.AiChatSessionMapper;
import com.example.FreStyle.repository.AiChatSessionRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetAiChatSessionsByRelatedRoomIdUseCase")
class GetAiChatSessionsByRelatedRoomIdUseCaseTest {

    @Mock private AiChatSessionRepository repository;
    @Mock private AiChatSessionMapper mapper;
    @InjectMocks private GetAiChatSessionsByRelatedRoomIdUseCase useCase;

    private AiChatSession createSession(Integer id) {
        AiChatSession session = new AiChatSession();
        session.setId(id);
        User user = new User();
        user.setId(1);
        session.setUser(user);
        session.setTitle("セッション" + id);
        return session;
    }

    @Test
    @DisplayName("ルームIDでセッション一覧を正常に取得できる")
    void returnsSessions() {
        AiChatSession s1 = createSession(1);
        AiChatSession s2 = createSession(2);
        AiChatSessionDto dto1 = new AiChatSessionDto(1, null, null, null, null, null, null, null, null);
        AiChatSessionDto dto2 = new AiChatSessionDto(2, null, null, null, null, null, null, null, null);

        when(repository.findByRelatedRoomId(10)).thenReturn(List.of(s1, s2));
        when(mapper.toDto(s1)).thenReturn(dto1);
        when(mapper.toDto(s2)).thenReturn(dto2);

        List<AiChatSessionDto> result = useCase.execute(10);

        assertThat(result).hasSize(2);
        assertThat(result.get(0).id()).isEqualTo(1);
        assertThat(result.get(1).id()).isEqualTo(2);
        verify(repository).findByRelatedRoomId(10);
    }

    @Test
    @DisplayName("セッションが0件の場合は空リストを返す")
    void returnsEmptyList() {
        when(repository.findByRelatedRoomId(99)).thenReturn(Collections.emptyList());

        List<AiChatSessionDto> result = useCase.execute(99);

        assertThat(result).isEmpty();
    }
}
