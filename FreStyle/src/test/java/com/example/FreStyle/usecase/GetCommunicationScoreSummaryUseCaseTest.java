package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.verify;
import static org.mockito.Mockito.when;

import java.sql.Timestamp;
import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.CommunicationScoreSummaryDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.CommunicationScoreRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetCommunicationScoreSummaryUseCase テスト")
class GetCommunicationScoreSummaryUseCaseTest {

    @Mock
    private CommunicationScoreRepository communicationScoreRepository;

    @InjectMocks
    private GetCommunicationScoreSummaryUseCase useCase;

    private CommunicationScore score(Integer sessionId, String axisName, int scoreVal) {
        CommunicationScore cs = new CommunicationScore();
        AiChatSession session = new AiChatSession();
        session.setId(sessionId);
        cs.setSession(session);
        User user = new User();
        user.setId(1);
        cs.setUser(user);
        cs.setAxisName(axisName);
        cs.setScore(scoreVal);
        cs.setCreatedAt(new Timestamp(System.currentTimeMillis()));
        return cs;
    }

    @Test
    @DisplayName("複数セッション・複数軸のスコアサマリーを集計する")
    void execute_returnsCorrectSummary() {
        List<CommunicationScore> scores = List.of(
                score(1, "論理的構成力", 8),
                score(1, "配慮表現", 6),
                score(2, "論理的構成力", 10),
                score(2, "配慮表現", 4)
        );

        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1)).thenReturn(scores);

        CommunicationScoreSummaryDto result = useCase.execute(1);

        assertThat(result.totalSessions()).isEqualTo(2);
        assertThat(result.overallAverage()).isEqualTo(7.0);
        assertThat(result.axisAverages()).hasSize(2);
        assertThat(result.bestAxis()).isEqualTo("論理的構成力");
        assertThat(result.worstAxis()).isEqualTo("配慮表現");
        verify(communicationScoreRepository).findByUserIdOrderByCreatedAtDesc(1);
    }

    @Test
    @DisplayName("スコアが空の場合はゼロサマリーを返す")
    void execute_emptyScores_returnsZero() {
        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1)).thenReturn(List.of());

        CommunicationScoreSummaryDto result = useCase.execute(1);

        assertThat(result.totalSessions()).isZero();
        assertThat(result.overallAverage()).isEqualTo(0.0);
        assertThat(result.axisAverages()).isEmpty();
        assertThat(result.bestAxis()).isNull();
        assertThat(result.worstAxis()).isNull();
        verify(communicationScoreRepository).findByUserIdOrderByCreatedAtDesc(1);
    }

    @Test
    @DisplayName("1セッション・1軸の場合はbestとworstが同じ")
    void execute_singleAxis_bestAndWorstSame() {
        List<CommunicationScore> scores = List.of(
                score(1, "要約力", 7)
        );

        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1)).thenReturn(scores);

        CommunicationScoreSummaryDto result = useCase.execute(1);

        assertThat(result.totalSessions()).isEqualTo(1);
        assertThat(result.overallAverage()).isEqualTo(7.0);
        assertThat(result.axisAverages()).hasSize(1);
        assertThat(result.axisAverages().get(0).axisName()).isEqualTo("要約力");
        assertThat(result.axisAverages().get(0).average()).isEqualTo(7.0);
        assertThat(result.axisAverages().get(0).count()).isEqualTo(1);
        assertThat(result.bestAxis()).isEqualTo("要約力");
        assertThat(result.worstAxis()).isEqualTo("要約力");
        verify(communicationScoreRepository).findByUserIdOrderByCreatedAtDesc(1);
    }

    @Test
    @DisplayName("軸別平均が降順でソートされる")
    void execute_axisAveragesSortedDescending() {
        List<CommunicationScore> scores = List.of(
                score(1, "配慮表現", 5),
                score(1, "論理的構成力", 9),
                score(1, "要約力", 7)
        );

        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1)).thenReturn(scores);

        CommunicationScoreSummaryDto result = useCase.execute(1);

        assertThat(result.axisAverages().get(0).axisName()).isEqualTo("論理的構成力");
        assertThat(result.axisAverages().get(1).axisName()).isEqualTo("要約力");
        assertThat(result.axisAverages().get(2).axisName()).isEqualTo("配慮表現");
        verify(communicationScoreRepository).findByUserIdOrderByCreatedAtDesc(1);
    }
}
