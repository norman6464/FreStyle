package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.AiChatSessionStatsDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.repository.AiChatSessionRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetAiChatSessionStatsUseCase テスト")
class GetAiChatSessionStatsUseCaseTest {

    @Mock
    private AiChatSessionRepository aiChatSessionRepository;

    @InjectMocks
    private GetAiChatSessionStatsUseCase useCase;

    private AiChatSession session(String sessionType, String scene) {
        AiChatSession s = new AiChatSession();
        s.setSessionType(sessionType);
        s.setScene(scene);
        return s;
    }

    @Test
    @DisplayName("タイプ別・シーン別のセッション統計を集計する")
    void execute_returnsCorrectStats() {
        List<AiChatSession> sessions = List.of(
                session("normal", "meeting"),
                session("normal", "email"),
                session("practice", "meeting"),
                session("normal", "meeting")
        );

        when(aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(1)).thenReturn(sessions);

        AiChatSessionStatsDto result = useCase.execute(1);

        assertThat(result.totalSessions()).isEqualTo(4);
        assertThat(result.sessionsByType()).hasSize(2);
        assertThat(result.sessionsByScene()).hasSize(2);
        verify(aiChatSessionRepository).findByUserIdOrderByCreatedAtDesc(1);
    }

    @Test
    @DisplayName("セッションが空の場合はゼロ統計を返す")
    void execute_emptyReturnsZero() {
        when(aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(1)).thenReturn(List.of());

        AiChatSessionStatsDto result = useCase.execute(1);

        assertThat(result.totalSessions()).isZero();
        assertThat(result.sessionsByType()).isEmpty();
        assertThat(result.sessionsByScene()).isEmpty();
        verify(aiChatSessionRepository).findByUserIdOrderByCreatedAtDesc(1);
    }

    @Test
    @DisplayName("タイプ別カウントがカウント降順でソートされる")
    void execute_typesSortedByCountDescending() {
        List<AiChatSession> sessions = List.of(
                session("normal", null),
                session("normal", null),
                session("normal", null),
                session("practice", null)
        );

        when(aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(1)).thenReturn(sessions);

        AiChatSessionStatsDto result = useCase.execute(1);

        assertThat(result.sessionsByType().getFirst().type()).isEqualTo("normal");
        assertThat(result.sessionsByType().getFirst().count()).isEqualTo(3);
        assertThat(result.sessionsByType().getLast().count()).isEqualTo(1);
        verify(aiChatSessionRepository).findByUserIdOrderByCreatedAtDesc(1);
    }

    @Test
    @DisplayName("シーンがnullのセッションは除外される")
    void execute_nullScenesExcluded() {
        List<AiChatSession> sessions = List.of(
                session("normal", "meeting"),
                session("normal", null),
                session("practice", "email")
        );

        when(aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(1)).thenReturn(sessions);

        AiChatSessionStatsDto result = useCase.execute(1);

        assertThat(result.totalSessions()).isEqualTo(3);
        assertThat(result.sessionsByScene()).hasSize(2);
        assertThat(result.sessionsByScene().stream()
                .noneMatch(sc -> sc.scene() == null)).isTrue();
        verify(aiChatSessionRepository).findByUserIdOrderByCreatedAtDesc(1);
    }
}
