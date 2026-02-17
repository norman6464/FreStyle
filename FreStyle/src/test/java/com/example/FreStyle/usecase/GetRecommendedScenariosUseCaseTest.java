package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.*;

import java.util.List;
import java.util.Optional;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.RecommendedScenarioDto;
import com.example.FreStyle.dto.RecommendedScenarioDto.ScenarioRecommendation;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.entity.PracticeScenario;
import com.example.FreStyle.repository.AiChatSessionRepository;
import com.example.FreStyle.repository.CommunicationScoreRepository;
import com.example.FreStyle.repository.PracticeScenarioRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetRecommendedScenariosUseCase")
class GetRecommendedScenariosUseCaseTest {

    @Mock
    private AiChatSessionRepository aiChatSessionRepository;

    @Mock
    private CommunicationScoreRepository communicationScoreRepository;

    @Mock
    private PracticeScenarioRepository practiceScenarioRepository;

    @InjectMocks
    private GetRecommendedScenariosUseCase useCase;

    private AiChatSession practiceSession(Integer id, Integer scenarioId) {
        AiChatSession session = new AiChatSession();
        session.setId(id);
        session.setScenarioId(scenarioId);
        session.setSessionType("practice");
        return session;
    }

    private AiChatSession normalSession(Integer id) {
        AiChatSession session = new AiChatSession();
        session.setId(id);
        session.setSessionType("normal");
        return session;
    }

    private CommunicationScore score(AiChatSession session, String axis, int scoreValue) {
        CommunicationScore cs = new CommunicationScore();
        cs.setSession(session);
        cs.setAxisName(axis);
        cs.setScore(scoreValue);
        return cs;
    }

    private PracticeScenario scenario(Integer id, String name, String category, String difficulty) {
        PracticeScenario ps = new PracticeScenario();
        ps.setId(id);
        ps.setName(name);
        ps.setCategory(category);
        ps.setDifficulty(difficulty);
        return ps;
    }

    @Test
    @DisplayName("練習セッションに基づく推奨シナリオを返す")
    void returnsRecommendedScenarios() {
        AiChatSession s1 = practiceSession(1, 10);
        AiChatSession s2 = practiceSession(2, 20);

        when(aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of(s1, s2));
        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of(
                        score(s1, "論理性", 6), score(s1, "共感力", 8),
                        score(s2, "論理性", 4), score(s2, "共感力", 5)));

        PracticeScenario ps10 = scenario(10, "会議シナリオ", "business", "easy");
        PracticeScenario ps20 = scenario(20, "交渉シナリオ", "business", "hard");
        when(practiceScenarioRepository.findById(10)).thenReturn(Optional.of(ps10));
        when(practiceScenarioRepository.findById(20)).thenReturn(Optional.of(ps20));

        RecommendedScenarioDto result = useCase.execute(1);

        assertThat(result.recommendations()).hasSize(2);
        // スコアが低い順
        assertThat(result.recommendations().get(0).scenarioName()).isEqualTo("交渉シナリオ");
        assertThat(result.recommendations().get(0).averageScore()).isEqualTo(4.5);
        assertThat(result.recommendations().get(1).scenarioName()).isEqualTo("会議シナリオ");
        assertThat(result.recommendations().get(1).averageScore()).isEqualTo(7.0);
    }

    @Test
    @DisplayName("練習履歴がない場合は空リストを返す")
    void returnsEmptyWhenNoPracticeSessions() {
        AiChatSession normal = normalSession(1);

        when(aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of(normal));

        RecommendedScenarioDto result = useCase.execute(1);

        assertThat(result.recommendations()).isEmpty();
    }

    @Test
    @DisplayName("スコアデータがないシナリオは除外される")
    void excludesScenariosWithoutScores() {
        AiChatSession s1 = practiceSession(1, 10);
        AiChatSession s2 = practiceSession(2, 20);

        when(aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of(s1, s2));
        // s2(scenarioId=20)にはスコアなし
        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of(score(s1, "論理性", 7)));

        PracticeScenario ps10 = scenario(10, "会議シナリオ", "business", "easy");
        when(practiceScenarioRepository.findById(10)).thenReturn(Optional.of(ps10));

        RecommendedScenarioDto result = useCase.execute(1);

        assertThat(result.recommendations()).hasSize(1);
        assertThat(result.recommendations().get(0).scenarioId()).isEqualTo(10);
    }

    @Test
    @DisplayName("同一シナリオの複数セッションのスコアを集約する")
    void aggregatesScoresAcrossMultipleSessions() {
        AiChatSession s1 = practiceSession(1, 10);
        AiChatSession s2 = practiceSession(2, 10);

        when(aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of(s1, s2));
        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of(
                        score(s1, "論理性", 6), score(s1, "共感力", 8),
                        score(s2, "論理性", 4), score(s2, "共感力", 6)));

        PracticeScenario ps10 = scenario(10, "会議シナリオ", "business", "easy");
        when(practiceScenarioRepository.findById(10)).thenReturn(Optional.of(ps10));

        RecommendedScenarioDto result = useCase.execute(1);

        assertThat(result.recommendations()).hasSize(1);
        assertThat(result.recommendations().get(0).averageScore()).isEqualTo(6.0);
        assertThat(result.recommendations().get(0).practiceCount()).isEqualTo(2);
    }

    @Test
    @DisplayName("最大5件に制限される")
    void limitsToFiveRecommendations() {
        List<AiChatSession> sessions = new java.util.ArrayList<>();
        List<CommunicationScore> scores = new java.util.ArrayList<>();

        for (int i = 1; i <= 7; i++) {
            AiChatSession s = practiceSession(i, i * 10);
            sessions.add(s);
            scores.add(score(s, "論理性", i));
            PracticeScenario ps = scenario(i * 10, "シナリオ" + i, "business", "easy");
            when(practiceScenarioRepository.findById(i * 10)).thenReturn(Optional.of(ps));
        }

        when(aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(sessions);
        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(scores);

        RecommendedScenarioDto result = useCase.execute(1);

        assertThat(result.recommendations()).hasSize(5);
    }
}
