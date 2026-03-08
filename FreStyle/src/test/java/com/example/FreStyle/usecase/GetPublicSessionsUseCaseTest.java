package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.when;

import java.sql.Timestamp;
import java.util.List;

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
import com.example.FreStyle.repository.SharedSessionRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetPublicSessionsUseCase")
class GetPublicSessionsUseCaseTest {

    @Mock
    private SharedSessionRepository sharedSessionRepository;

    @InjectMocks
    private GetPublicSessionsUseCase useCase;

    @Test
    @DisplayName("公開セッションを取得する")
    void shouldReturnPublicSessions() {
        User user = new User();
        user.setId(1);
        user.setName("ユーザー1");
        user.setIconUrl("https://example.com/icon.png");

        AiChatSession session = new AiChatSession();
        session.setId(10);
        session.setUser(user);
        session.setTitle("公開セッション");

        SharedSession shared = new SharedSession();
        shared.setId(100);
        shared.setSession(session);
        shared.setUser(user);
        shared.setDescription("説明");
        shared.setIsPublic(true);
        shared.setCreatedAt(Timestamp.valueOf("2026-03-08 10:00:00"));

        when(sharedSessionRepository.findPublicSessions()).thenReturn(List.of(shared));

        List<SharedSessionDto> result = useCase.execute();

        assertThat(result).hasSize(1);
        assertThat(result.get(0).sessionTitle()).isEqualTo("公開セッション");
        assertThat(result.get(0).username()).isEqualTo("ユーザー1");
    }
}
