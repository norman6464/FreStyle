package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.*;

import java.util.Optional;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.AiChatSessionDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.mapper.AiChatSessionMapper;
import com.example.FreStyle.repository.AiChatSessionRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetAiChatSessionByIdUseCase")
class GetAiChatSessionByIdUseCaseTest {

    @Mock private AiChatSessionRepository repository;
    @Mock private AiChatSessionMapper mapper;
    @InjectMocks private GetAiChatSessionByIdUseCase useCase;

    private AiChatSession createSession(Integer id, Integer userId) {
        AiChatSession session = new AiChatSession();
        session.setId(id);
        User user = new User();
        user.setId(userId);
        session.setUser(user);
        session.setTitle("テストセッション");
        return session;
    }

    @Test
    @DisplayName("セッションを正常に取得できる")
    void returnsSession() {
        AiChatSession session = createSession(1, 10);
        AiChatSessionDto dto = new AiChatSessionDto(1, null, null, null, null, null, null, null, null);
        when(repository.findByIdAndUserId(1, 10)).thenReturn(Optional.of(session));
        when(mapper.toDto(session)).thenReturn(dto);

        AiChatSessionDto result = useCase.execute(1, 10);

        assertThat(result.id()).isEqualTo(1);
        verify(repository).findByIdAndUserId(1, 10);
    }

    @Test
    @DisplayName("セッションが見つからない場合はResourceNotFoundExceptionを投げる")
    void throwsWhenNotFound() {
        when(repository.findByIdAndUserId(999, 10)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> useCase.execute(999, 10))
            .isInstanceOf(ResourceNotFoundException.class)
            .hasMessageContaining("セッションが見つからないか");
    }

    @Test
    @DisplayName("別ユーザーのセッションにはアクセスできない")
    void throwsWhenUnauthorized() {
        when(repository.findByIdAndUserId(1, 99)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> useCase.execute(1, 99))
            .isInstanceOf(ResourceNotFoundException.class);
    }
}
