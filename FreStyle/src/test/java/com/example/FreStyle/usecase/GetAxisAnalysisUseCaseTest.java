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

import com.example.FreStyle.dto.AxisAnalysisDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.CommunicationScoreRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetAxisAnalysisUseCase テスト")
class GetAxisAnalysisUseCaseTest {

    @Mock
    private CommunicationScoreRepository communicationScoreRepository;

    @InjectMocks
    private GetAxisAnalysisUseCase useCase;

    @Test
    @DisplayName("複数軸のスコアが正しく集計される")
    void multipleAxesAreAggregatedCorrectly() {
        List<CommunicationScore> scores = List.of(
                createScore(1, "論理的構成力", 8),
                createScore(1, "論理的構成力", 6),
                createScore(1, "傾聴力", 9),
                createScore(1, "傾聴力", 7),
                createScore(1, "表現力", 5));

        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(scores);

        AxisAnalysisDto result = useCase.execute(1);

        assertThat(result.axisStats()).hasSize(3);
        assertThat(result.totalEvaluations()).isEqualTo(5);
        // 平均スコア降順で並んでいることを検証
        assertThat(result.axisStats().get(0).axisName()).isEqualTo("傾聴力"); // avg 8.0
        assertThat(result.axisStats().get(1).axisName()).isEqualTo("論理的構成力"); // avg 7.0
        assertThat(result.axisStats().get(2).axisName()).isEqualTo("表現力"); // avg 5.0
        assertThat(result.bestAxis()).isEqualTo("傾聴力");
        assertThat(result.worstAxis()).isEqualTo("表現力");
    }

    @Test
    @DisplayName("各軸の平均スコアが正しく計算される")
    void averageScoreIsCalculatedPerAxis() {
        List<CommunicationScore> scores = List.of(
                createScore(1, "論理的構成力", 8),
                createScore(1, "論理的構成力", 6));

        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(scores);

        AxisAnalysisDto result = useCase.execute(1);

        AxisAnalysisDto.AxisStat stat = result.axisStats().stream()
                .filter(s -> s.axisName().equals("論理的構成力"))
                .findFirst().orElseThrow();
        assertThat(stat.averageScore()).isEqualTo(7.0);
        assertThat(stat.count()).isEqualTo(2);
    }

    @Test
    @DisplayName("得意軸と苦手軸が正しく判定される")
    void bestAndWorstAxesAreIdentified() {
        List<CommunicationScore> scores = List.of(
                createScore(1, "傾聴力", 9),
                createScore(1, "論理的構成力", 5),
                createScore(1, "表現力", 7));

        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(scores);

        AxisAnalysisDto result = useCase.execute(1);

        assertThat(result.bestAxis()).isEqualTo("傾聴力");
        assertThat(result.worstAxis()).isEqualTo("論理的構成力");
    }

    @Test
    @DisplayName("スコアが0件の場合は空の分析結果を返す")
    void emptyScoresReturnEmptyAnalysis() {
        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(List.of());

        AxisAnalysisDto result = useCase.execute(1);

        assertThat(result.axisStats()).isEmpty();
        assertThat(result.bestAxis()).isNull();
        assertThat(result.worstAxis()).isNull();
        assertThat(result.totalEvaluations()).isEqualTo(0);
    }

    @Test
    @DisplayName("軸が1つだけの場合はbestとworstが同じになる")
    void singleAxisIsBothBestAndWorst() {
        List<CommunicationScore> scores = List.of(
                createScore(1, "傾聴力", 8),
                createScore(1, "傾聴力", 6));

        when(communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(1))
                .thenReturn(scores);

        AxisAnalysisDto result = useCase.execute(1);

        assertThat(result.bestAxis()).isEqualTo("傾聴力");
        assertThat(result.worstAxis()).isEqualTo("傾聴力");
        assertThat(result.axisStats()).hasSize(1);
    }

    private CommunicationScore createScore(Integer userId, String axisName, int score) {
        CommunicationScore cs = new CommunicationScore();
        User user = new User();
        user.setId(userId);
        cs.setUser(user);
        AiChatSession session = new AiChatSession();
        session.setId(1);
        cs.setSession(session);
        cs.setAxisName(axisName);
        cs.setScore(score);
        cs.setCreatedAt(new Timestamp(System.currentTimeMillis()));
        return cs;
    }
}
