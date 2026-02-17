package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.sql.Timestamp;
import java.util.List;
import java.util.Optional;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Nested;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.PracticeSessionSummaryDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.entity.PracticeScenario;
import com.example.FreStyle.entity.SessionNote;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.exception.ResourceNotFoundException;
import com.example.FreStyle.repository.AiChatMessageRepository;
import com.example.FreStyle.repository.AiChatSessionRepository;
import com.example.FreStyle.repository.CommunicationScoreRepository;
import com.example.FreStyle.repository.PracticeScenarioRepository;
import com.example.FreStyle.repository.SessionNoteRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetPracticeSessionSummaryUseCase")
class GetPracticeSessionSummaryUseCaseTest {

    @Mock private AiChatSessionRepository aiChatSessionRepository;
    @Mock private AiChatMessageRepository aiChatMessageRepository;
    @Mock private CommunicationScoreRepository communicationScoreRepository;
    @Mock private SessionNoteRepository sessionNoteRepository;
    @Mock private PracticeScenarioRepository practiceScenarioRepository;

    @InjectMocks
    private GetPracticeSessionSummaryUseCase useCase;

    @Nested
    @DisplayName("正常系")
    class Success {

        @Test
        @DisplayName("スコア・ノート・シナリオ付きのサマリーを取得する")
        void returnsFullSummary() {
            AiChatSession session = createSession(1, 10, "練習セッション", "practice", "meeting", 5);
            when(aiChatSessionRepository.findByIdAndUserId(1, 10)).thenReturn(Optional.of(session));
            when(aiChatMessageRepository.countBySessionId(1)).thenReturn(8L);

            CommunicationScore score1 = createScore("論理的構成力", 85, "良い構成");
            CommunicationScore score2 = createScore("配慮表現", 60, "改善の余地あり");
            when(communicationScoreRepository.findBySessionId(1)).thenReturn(List.of(score1, score2));

            SessionNote noteEntity = new SessionNote();
            noteEntity.setNote("要復習");
            when(sessionNoteRepository.findByUserIdAndSessionId(10, 1)).thenReturn(Optional.of(noteEntity));

            PracticeScenario scenario = new PracticeScenario();
            scenario.setName("本番障害の緊急報告");
            when(practiceScenarioRepository.findById(5)).thenReturn(Optional.of(scenario));

            PracticeSessionSummaryDto result = useCase.execute(1, 10);

            assertThat(result.sessionId()).isEqualTo(1);
            assertThat(result.title()).isEqualTo("練習セッション");
            assertThat(result.sessionType()).isEqualTo("practice");
            assertThat(result.messageCount()).isEqualTo(8L);
            assertThat(result.scores()).hasSize(2);
            assertThat(result.averageScore()).isEqualTo(72.5);
            assertThat(result.bestAxis()).isEqualTo("論理的構成力");
            assertThat(result.worstAxis()).isEqualTo("配慮表現");
            assertThat(result.note()).isEqualTo("要復習");
            assertThat(result.scenarioName()).isEqualTo("本番障害の緊急報告");

            verify(aiChatSessionRepository).findByIdAndUserId(1, 10);
            verify(aiChatMessageRepository).countBySessionId(1);
            verify(communicationScoreRepository).findBySessionId(1);
            verify(sessionNoteRepository).findByUserIdAndSessionId(10, 1);
            verify(practiceScenarioRepository).findById(5);
        }

        @Test
        @DisplayName("スコアがないセッションのサマリーを取得する")
        void returnsWithoutScores() {
            AiChatSession session = createSession(2, 10, "通常セッション", "normal", null, null);
            when(aiChatSessionRepository.findByIdAndUserId(2, 10)).thenReturn(Optional.of(session));
            when(aiChatMessageRepository.countBySessionId(2)).thenReturn(3L);
            when(communicationScoreRepository.findBySessionId(2)).thenReturn(List.of());
            when(sessionNoteRepository.findByUserIdAndSessionId(10, 2)).thenReturn(Optional.empty());

            PracticeSessionSummaryDto result = useCase.execute(2, 10);

            assertThat(result.sessionId()).isEqualTo(2);
            assertThat(result.scores()).isEmpty();
            assertThat(result.averageScore()).isNull();
            assertThat(result.bestAxis()).isNull();
            assertThat(result.worstAxis()).isNull();
            assertThat(result.note()).isNull();
            assertThat(result.scenarioName()).isNull();
        }

        @Test
        @DisplayName("スコアが1軸のみの場合はworstAxisがnullになる")
        void singleScoreAxis() {
            AiChatSession session = createSession(3, 10, "テスト", "practice", null, null);
            when(aiChatSessionRepository.findByIdAndUserId(3, 10)).thenReturn(Optional.of(session));
            when(aiChatMessageRepository.countBySessionId(3)).thenReturn(5L);

            CommunicationScore score = createScore("論理的構成力", 90, null);
            when(communicationScoreRepository.findBySessionId(3)).thenReturn(List.of(score));
            when(sessionNoteRepository.findByUserIdAndSessionId(10, 3)).thenReturn(Optional.empty());

            PracticeSessionSummaryDto result = useCase.execute(3, 10);

            assertThat(result.bestAxis()).isEqualTo("論理的構成力");
            assertThat(result.worstAxis()).isNull();
            assertThat(result.averageScore()).isEqualTo(90.0);
        }

        @Test
        @DisplayName("シナリオが削除されていてもサマリーを取得できる")
        void scenarioDeleted() {
            AiChatSession session = createSession(4, 10, "テスト", "practice", null, 99);
            when(aiChatSessionRepository.findByIdAndUserId(4, 10)).thenReturn(Optional.of(session));
            when(aiChatMessageRepository.countBySessionId(4)).thenReturn(2L);
            when(communicationScoreRepository.findBySessionId(4)).thenReturn(List.of());
            when(sessionNoteRepository.findByUserIdAndSessionId(10, 4)).thenReturn(Optional.empty());
            when(practiceScenarioRepository.findById(99)).thenReturn(Optional.empty());

            PracticeSessionSummaryDto result = useCase.execute(4, 10);

            assertThat(result.scenarioName()).isNull();
        }

        @Test
        @DisplayName("全スコアが同じ値の場合はworstAxisがnullになる")
        void equalScores_worstAxisIsNull() {
            AiChatSession session = createSession(5, 10, "同一スコア", "practice", null, null);
            when(aiChatSessionRepository.findByIdAndUserId(5, 10)).thenReturn(Optional.of(session));
            when(aiChatMessageRepository.countBySessionId(5)).thenReturn(0L);

            CommunicationScore s1 = createScore("話し方", 80, null);
            CommunicationScore s2 = createScore("内容", 80, null);
            when(communicationScoreRepository.findBySessionId(5)).thenReturn(List.of(s1, s2));
            when(sessionNoteRepository.findByUserIdAndSessionId(10, 5)).thenReturn(Optional.empty());

            PracticeSessionSummaryDto result = useCase.execute(5, 10);

            assertThat(result.bestAxis()).isNotNull();
            assertThat(result.worstAxis()).isNull();
            assertThat(result.averageScore()).isEqualTo(80.0);
        }
    }

    @Nested
    @DisplayName("異常系")
    class Error {

        @Test
        @DisplayName("存在しないセッションの場合はResourceNotFoundExceptionをスローする")
        void throwsWhenSessionNotFound() {
            when(aiChatSessionRepository.findByIdAndUserId(999, 10)).thenReturn(Optional.empty());

            assertThatThrownBy(() -> useCase.execute(999, 10))
                    .isInstanceOf(ResourceNotFoundException.class);
        }
    }

    private AiChatSession createSession(Integer id, Integer userId, String title, String sessionType, String scene, Integer scenarioId) {
        AiChatSession session = new AiChatSession();
        session.setId(id);
        User user = new User();
        user.setId(userId);
        session.setUser(user);
        session.setTitle(title);
        session.setSessionType(sessionType);
        session.setScene(scene);
        session.setScenarioId(scenarioId);
        session.setCreatedAt(new Timestamp(System.currentTimeMillis()));
        return session;
    }

    private CommunicationScore createScore(String axisName, Integer score, String comment) {
        CommunicationScore cs = new CommunicationScore();
        cs.setAxisName(axisName);
        cs.setScore(score);
        cs.setComment(comment);
        return cs;
    }
}
