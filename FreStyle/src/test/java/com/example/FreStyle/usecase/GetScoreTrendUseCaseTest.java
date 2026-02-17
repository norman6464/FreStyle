package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.when;

import java.sql.Timestamp;
import java.time.LocalDateTime;
import java.util.Collections;
import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.ScoreTrendDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.CommunicationScoreRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetScoreTrendUseCase テスト")
class GetScoreTrendUseCaseTest {

    @Mock
    private CommunicationScoreRepository communicationScoreRepository;

    @InjectMocks
    private GetScoreTrendUseCase useCase;

    private CommunicationScore createScore(int sessionId, int score, String axisName, LocalDateTime createdAt) {
        CommunicationScore cs = new CommunicationScore();
        AiChatSession session = new AiChatSession();
        session.setId(sessionId);
        cs.setSession(session);
        cs.setScore(score);
        cs.setAxisName(axisName);
        cs.setCreatedAt(Timestamp.valueOf(createdAt));
        return cs;
    }

    @Test
    @DisplayName("複数セッションのスコアトレンドを取得する")
    void returnsScoreTrendForMultipleSessions() {
        LocalDateTime now = LocalDateTime.now();
        List<CommunicationScore> scores = List.of(
                createScore(1, 80, "明瞭性", now.minusDays(5)),
                createScore(1, 70, "論理性", now.minusDays(5)),
                createScore(2, 90, "明瞭性", now.minusDays(2)),
                createScore(2, 85, "論理性", now.minusDays(2)));

        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(scores);

        ScoreTrendDto result = useCase.execute(1, 30);

        assertThat(result.totalSessions()).isEqualTo(2);
        assertThat(result.sessionScores()).hasSize(2);
        assertThat(result.overallAverage()).isCloseTo(81.25, org.assertj.core.data.Offset.offset(0.01));
        assertThat(result.bestSession().sessionId()).isEqualTo(2);
        assertThat(result.bestSession().averageScore()).isCloseTo(87.5, org.assertj.core.data.Offset.offset(0.01));
    }

    @Test
    @DisplayName("スコアがない場合は空のトレンドを返す")
    void returnsEmptyTrendWhenNoScores() {
        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(Collections.emptyList());

        ScoreTrendDto result = useCase.execute(1, 30);

        assertThat(result.totalSessions()).isZero();
        assertThat(result.sessionScores()).isEmpty();
        assertThat(result.overallAverage()).isZero();
        assertThat(result.bestSession()).isNull();
    }

    @Test
    @DisplayName("期間外のスコアは除外される")
    void excludesScoresOutsidePeriod() {
        LocalDateTime now = LocalDateTime.now();
        List<CommunicationScore> scores = List.of(
                createScore(1, 80, "明瞭性", now.minusDays(5)),
                createScore(2, 90, "明瞭性", now.minusDays(60)));

        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(scores);

        ScoreTrendDto result = useCase.execute(1, 30);

        assertThat(result.totalSessions()).isEqualTo(1);
        assertThat(result.sessionScores()).hasSize(1);
        assertThat(result.sessionScores().get(0).sessionId()).isEqualTo(1);
    }

    @Test
    @DisplayName("セッションスコアは日付の古い順にソートされる")
    void sessionScoresAreSortedByDateAscending() {
        LocalDateTime now = LocalDateTime.now();
        List<CommunicationScore> scores = List.of(
                createScore(2, 90, "明瞭性", now.minusDays(1)),
                createScore(1, 80, "明瞭性", now.minusDays(10)));

        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(scores);

        ScoreTrendDto result = useCase.execute(1, 30);

        assertThat(result.sessionScores().get(0).sessionId()).isEqualTo(1);
        assertThat(result.sessionScores().get(1).sessionId()).isEqualTo(2);
    }

    @Test
    @DisplayName("days指定が反映される")
    void respectsDaysParameter() {
        LocalDateTime now = LocalDateTime.now();
        List<CommunicationScore> scores = List.of(
                createScore(1, 80, "明瞭性", now.minusDays(5)),
                createScore(2, 90, "明瞭性", now.minusDays(10)));

        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(scores);

        ScoreTrendDto result = useCase.execute(1, 7);

        assertThat(result.days()).isEqualTo(7);
        assertThat(result.totalSessions()).isEqualTo(1);
    }
}
