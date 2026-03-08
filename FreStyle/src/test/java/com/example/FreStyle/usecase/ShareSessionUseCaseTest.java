package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.sql.Timestamp;
import java.util.Optional;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.SharedSessionDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.SharedSession;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.form.ShareSessionForm;
import com.example.FreStyle.repository.AiChatSessionRepository;
import com.example.FreStyle.repository.SharedSessionRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("ShareSessionUseCase")
class ShareSessionUseCaseTest {

    @Mock
    private AiChatSessionRepository aiChatSessionRepository;

    @Mock
    private SharedSessionRepository sharedSessionRepository;

    @InjectMocks
    private ShareSessionUseCase useCase;

    @Test
    @DisplayName("セッションを共有する")
    void shouldShareSession() {
        User user = new User();
        user.setId(1);
        user.setName("テストユーザー");
        user.setIconUrl("https://example.com/icon.png");

        AiChatSession session = new AiChatSession();
        session.setId(10);
        session.setUser(user);
        session.setTitle("テストセッション");

        when(aiChatSessionRepository.findByIdAndUserId(10, 1))
                .thenReturn(Optional.of(session));

        SharedSession saved = new SharedSession();
        saved.setId(100);
        saved.setSession(session);
        saved.setUser(user);
        saved.setDescription("共有の説明");
        saved.setIsPublic(true);
        saved.setCreatedAt(Timestamp.valueOf("2026-03-08 10:00:00"));
        when(sharedSessionRepository.save(any(SharedSession.class))).thenReturn(saved);

        ShareSessionForm form = new ShareSessionForm(10, "共有の説明");
        SharedSessionDto result = useCase.execute(1, form);

        assertThat(result.id()).isEqualTo(100);
        assertThat(result.sessionId()).isEqualTo(10);
        assertThat(result.sessionTitle()).isEqualTo("テストセッション");
        assertThat(result.userId()).isEqualTo(1);
        assertThat(result.username()).isEqualTo("テストユーザー");
        assertThat(result.description()).isEqualTo("共有の説明");
        verify(sharedSessionRepository).save(any(SharedSession.class));
    }

    @Test
    @DisplayName("他人のセッションを共有しようとすると例外が発生する")
    void shouldThrowWhenSharingOthersSession() {
        when(aiChatSessionRepository.findByIdAndUserId(10, 2))
                .thenReturn(Optional.empty());

        ShareSessionForm form = new ShareSessionForm(10, "説明");

        assertThatThrownBy(() -> useCase.execute(2, form))
                .isInstanceOf(ResourceNotFoundException.class);
    }

    @Test
    @DisplayName("セッションが見つからない場合は例外を投げる")
    void shouldThrowWhenSessionNotFound() {
        when(aiChatSessionRepository.findByIdAndUserId(999, 1))
                .thenReturn(Optional.empty());

        ShareSessionForm form = new ShareSessionForm(999, "説明");

        assertThatThrownBy(() -> useCase.execute(1, form))
                .isInstanceOf(ResourceNotFoundException.class);
    }
}
