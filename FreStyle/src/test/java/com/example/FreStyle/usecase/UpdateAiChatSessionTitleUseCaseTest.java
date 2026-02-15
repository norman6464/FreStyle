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
@DisplayName("UpdateAiChatSessionTitleUseCase")
class UpdateAiChatSessionTitleUseCaseTest {

    @Mock private AiChatSessionRepository repository;
    @Mock private AiChatSessionMapper mapper;
    @InjectMocks private UpdateAiChatSessionTitleUseCase useCase;

    private AiChatSession createSession(Integer id, Integer userId, String title) {
        AiChatSession session = new AiChatSession();
        session.setId(id);
        User user = new User();
        user.setId(userId);
        session.setUser(user);
        session.setTitle(title);
        return session;
    }

    @Test
    @DisplayName("タイトルを正常に更新できる")
    void updatesTitle() {
        AiChatSession session = createSession(1, 10, "旧タイトル");
        AiChatSession saved = createSession(1, 10, "新タイトル");
        AiChatSessionDto dto = new AiChatSessionDto();
        dto.setId(1);
        dto.setTitle("新タイトル");

        when(repository.findByIdAndUserId(1, 10)).thenReturn(Optional.of(session));
        when(repository.save(session)).thenReturn(saved);
        when(mapper.toDto(saved)).thenReturn(dto);

        AiChatSessionDto result = useCase.execute(1, 10, "新タイトル");

        assertThat(result.getTitle()).isEqualTo("新タイトル");
        verify(repository).save(session);
    }

    @Test
    @DisplayName("セッションが見つからない場合はResourceNotFoundExceptionを投げる")
    void throwsWhenNotFound() {
        when(repository.findByIdAndUserId(999, 10)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> useCase.execute(999, 10, "タイトル"))
            .isInstanceOf(ResourceNotFoundException.class)
            .hasMessageContaining("セッションが見つからないか");
    }

    @Test
    @DisplayName("別ユーザーのセッションは更新できない")
    void throwsWhenUnauthorized() {
        when(repository.findByIdAndUserId(1, 99)).thenReturn(Optional.empty());

        assertThatThrownBy(() -> useCase.execute(1, 99, "タイトル"))
            .isInstanceOf(ResourceNotFoundException.class);
    }

    @Test
    @DisplayName("saveが呼ばれた時にタイトルが設定されている")
    void titleIsSetBeforeSave() {
        AiChatSession session = createSession(1, 10, "旧タイトル");
        AiChatSessionDto dto = new AiChatSessionDto();

        when(repository.findByIdAndUserId(1, 10)).thenReturn(Optional.of(session));
        when(repository.save(any())).thenReturn(session);
        when(mapper.toDto(any())).thenReturn(dto);

        useCase.execute(1, 10, "更新タイトル");

        assertThat(session.getTitle()).isEqualTo("更新タイトル");
    }
}
