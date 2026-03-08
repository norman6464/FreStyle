package com.example.FreStyle.service;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;

import java.sql.Timestamp;
import java.util.List;
import java.util.Optional;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.ArgumentCaptor;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.RankingDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.CommunicationScoreRepository;
import com.example.FreStyle.repository.UserRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("RankingService")
class RankingServiceTest {

    @Mock
    private CommunicationScoreRepository communicationScoreRepository;

    @Mock
    private UserRepository userRepository;

    @InjectMocks
    private RankingService rankingService;

    @Test
    @DisplayName("ユーザーの平均スコアでランキングを生成する")
    void shouldGenerateRankingByAverageScore() {
        // User1: avgScore=8.5, sessionCount=3
        // User2: avgScore=7.0, sessionCount=2
        Object[] row1 = new Object[]{1, 8.5, 3L};
        Object[] row2 = new Object[]{2, 7.0, 2L};
        when(communicationScoreRepository.findUserAverageScoresAfter(any(Timestamp.class)))
                .thenReturn(List.of(row1, row2));

        User user1 = new User();
        user1.setId(1);
        user1.setName("User1");
        user1.setIconUrl("icon1.png");
        when(userRepository.findById(1)).thenReturn(Optional.of(user1));

        User user2 = new User();
        user2.setId(2);
        user2.setName("User2");
        user2.setIconUrl("icon2.png");
        when(userRepository.findById(2)).thenReturn(Optional.of(user2));

        RankingDto result = rankingService.getRanking("weekly", 99);

        assertThat(result.entries()).hasSize(2);
        assertThat(result.entries().get(0).rank()).isEqualTo(1);
        assertThat(result.entries().get(0).username()).isEqualTo("User1");
        assertThat(result.entries().get(0).averageScore()).isEqualTo(8.5);
        assertThat(result.entries().get(1).rank()).isEqualTo(2);
        assertThat(result.entries().get(1).username()).isEqualTo("User2");
        assertThat(result.entries().get(1).averageScore()).isEqualTo(7.0);
        assertThat(result.myRanking()).isNull();
    }

    @Test
    @DisplayName("現在のユーザーのランキングをmyRankingに設定する")
    void shouldSetMyRankingForCurrentUser() {
        Object[] row1 = new Object[]{1, 9.0, 5L};
        Object[] row2 = new Object[]{2, 7.5, 3L};
        when(communicationScoreRepository.findUserAverageScoresAfter(any(Timestamp.class)))
                .thenReturn(List.of(row1, row2));

        User user1 = new User();
        user1.setId(1);
        user1.setName("User1");
        when(userRepository.findById(1)).thenReturn(Optional.of(user1));

        User user2 = new User();
        user2.setId(2);
        user2.setName("User2");
        when(userRepository.findById(2)).thenReturn(Optional.of(user2));

        RankingDto result = rankingService.getRanking("weekly", 2);

        assertThat(result.myRanking()).isNotNull();
        assertThat(result.myRanking().userId()).isEqualTo(2);
        assertThat(result.myRanking().rank()).isEqualTo(2);
        assertThat(result.myRanking().averageScore()).isEqualTo(7.5);
    }

    @Test
    @DisplayName("weeklyの場合は1週間前からのデータを取得する")
    void shouldUseCutoffForWeeklyPeriod() {
        ArgumentCaptor<Timestamp> cutoffCaptor = ArgumentCaptor.forClass(Timestamp.class);
        when(communicationScoreRepository.findUserAverageScoresAfter(cutoffCaptor.capture()))
                .thenReturn(List.of());

        rankingService.getRanking("weekly", 1);

        Timestamp capturedCutoff = cutoffCaptor.getValue();
        Timestamp now = new Timestamp(System.currentTimeMillis());
        // cutoff should be approximately 7 days ago
        long diffMillis = now.getTime() - capturedCutoff.getTime();
        long sevenDaysMillis = 7 * 24 * 60 * 60 * 1000L;
        // Allow 5 seconds tolerance
        assertThat(diffMillis).isBetween(sevenDaysMillis - 5000L, sevenDaysMillis + 5000L);
    }

    @Test
    @DisplayName("monthlyの場合は1ヶ月前からのデータを取得する")
    void shouldUseCutoffForMonthlyPeriod() {
        ArgumentCaptor<Timestamp> cutoffCaptor = ArgumentCaptor.forClass(Timestamp.class);
        when(communicationScoreRepository.findUserAverageScoresAfter(cutoffCaptor.capture()))
                .thenReturn(List.of());

        rankingService.getRanking("monthly", 1);

        Timestamp capturedCutoff = cutoffCaptor.getValue();
        Timestamp now = new Timestamp(System.currentTimeMillis());
        long diffMillis = now.getTime() - capturedCutoff.getTime();
        // approximately 30 days
        long thirtyDaysMillis = 30 * 24 * 60 * 60 * 1000L;
        // Allow range of 28-31 days to account for month length variation
        long twentyEightDaysMillis = 28 * 24 * 60 * 60 * 1000L;
        long thirtyOneDaysMillis = 31 * 24 * 60 * 60 * 1000L;
        assertThat(diffMillis).isBetween(twentyEightDaysMillis - 5000L, thirtyOneDaysMillis + 5000L);
    }

    @Test
    @DisplayName("ユーザーが見つからない場合はスキップする")
    void shouldSkipWhenUserNotFound() {
        Object[] row1 = new Object[]{1, 8.5, 3L};
        Object[] row2 = new Object[]{2, 7.0, 2L};
        when(communicationScoreRepository.findUserAverageScoresAfter(any(Timestamp.class)))
                .thenReturn(List.of(row1, row2));

        User user1 = new User();
        user1.setId(1);
        user1.setName("User1");
        user1.setIconUrl("icon1.png");
        when(userRepository.findById(1)).thenReturn(Optional.of(user1));
        when(userRepository.findById(2)).thenReturn(Optional.empty());

        RankingDto result = rankingService.getRanking("weekly", 99);

        assertThat(result.entries()).hasSize(1);
        assertThat(result.entries().get(0).username()).isEqualTo("User1");
    }

    @Test
    @DisplayName("空の結果の場合は空のリストを返す")
    void shouldReturnEmptyListWhenNoResults() {
        when(communicationScoreRepository.findUserAverageScoresAfter(any(Timestamp.class)))
                .thenReturn(List.of());

        RankingDto result = rankingService.getRanking("weekly", 1);

        assertThat(result.entries()).isEmpty();
        assertThat(result.myRanking()).isNull();
    }

    @Test
    @DisplayName("デフォルトの期間はmonthlyとして扱う")
    void shouldTreatUnknownPeriodAsMonthly() {
        ArgumentCaptor<Timestamp> cutoffCaptor = ArgumentCaptor.forClass(Timestamp.class);
        when(communicationScoreRepository.findUserAverageScoresAfter(cutoffCaptor.capture()))
                .thenReturn(List.of());

        rankingService.getRanking("unknown_period", 1);

        Timestamp capturedCutoff = cutoffCaptor.getValue();
        Timestamp now = new Timestamp(System.currentTimeMillis());
        long diffMillis = now.getTime() - capturedCutoff.getTime();
        // default falls through to monthly (approximately 28-31 days)
        long twentyEightDaysMillis = 28 * 24 * 60 * 60 * 1000L;
        long thirtyOneDaysMillis = 31 * 24 * 60 * 60 * 1000L;
        assertThat(diffMillis).isBetween(twentyEightDaysMillis - 5000L, thirtyOneDaysMillis + 5000L);
    }
}
