package com.example.FreStyle.usecase;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.when;

import java.time.LocalDate;
import java.util.List;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.DailyGoalStreakDto;
import com.example.FreStyle.entity.DailyGoal;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.DailyGoalRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetDailyGoalStreakUseCase テスト")
class GetDailyGoalStreakUseCaseTest {

    @Mock
    private DailyGoalRepository dailyGoalRepository;

    @InjectMocks
    private GetDailyGoalStreakUseCase getDailyGoalStreakUseCase;

    private User user() {
        User u = new User();
        u.setId(1);
        return u;
    }

    private DailyGoal goal(LocalDate date, int target, int completed) {
        return new DailyGoal(null, user(), date, target, completed);
    }

    @Test
    @DisplayName("連続達成日数を正しく計算する（今日を含む）")
    void execute_calculatesCurrentStreak() {
        LocalDate today = LocalDate.now();
        List<DailyGoal> goals = List.of(
                goal(today, 3, 3),
                goal(today.minusDays(1), 3, 5),
                goal(today.minusDays(2), 3, 3),
                goal(today.minusDays(3), 3, 1) // 未達成
        );

        when(dailyGoalRepository.findByUserIdOrderByGoalDateDesc(1)).thenReturn(goals);

        DailyGoalStreakDto result = getDailyGoalStreakUseCase.execute(1, today);

        assertThat(result.currentStreak()).isEqualTo(3);
        assertThat(result.longestStreak()).isEqualTo(3);
        assertThat(result.totalAchievedDays()).isEqualTo(3);
    }

    @Test
    @DisplayName("データがない場合は全て0を返す")
    void execute_noData_returnsZeros() {
        when(dailyGoalRepository.findByUserIdOrderByGoalDateDesc(1)).thenReturn(List.of());

        DailyGoalStreakDto result = getDailyGoalStreakUseCase.execute(1, LocalDate.now());

        assertThat(result.currentStreak()).isZero();
        assertThat(result.longestStreak()).isZero();
        assertThat(result.totalAchievedDays()).isZero();
    }

    @Test
    @DisplayName("今日が未達成でも昨日から連続していればcurrentStreakを計算する")
    void execute_todayNotAchieved_countsFromYesterday() {
        LocalDate today = LocalDate.now();
        List<DailyGoal> goals = List.of(
                goal(today, 3, 1),            // 今日は未達成
                goal(today.minusDays(1), 3, 3), // 昨日は達成
                goal(today.minusDays(2), 3, 4)  // 一昨日は達成
        );

        when(dailyGoalRepository.findByUserIdOrderByGoalDateDesc(1)).thenReturn(goals);

        DailyGoalStreakDto result = getDailyGoalStreakUseCase.execute(1, today);

        assertThat(result.currentStreak()).isEqualTo(2);
        assertThat(result.totalAchievedDays()).isEqualTo(2);
    }

    @Test
    @DisplayName("過去の最長ストリークがcurrentStreakより大きい場合")
    void execute_longestStreakGreaterThanCurrent() {
        LocalDate today = LocalDate.now();
        List<DailyGoal> goals = List.of(
                goal(today, 3, 3),               // 今日達成
                goal(today.minusDays(1), 3, 1),   // 昨日未達成（ストリーク切れ）
                goal(today.minusDays(2), 3, 3),   // 達成
                goal(today.minusDays(3), 3, 3),   // 達成
                goal(today.minusDays(4), 3, 3)    // 達成
        );

        when(dailyGoalRepository.findByUserIdOrderByGoalDateDesc(1)).thenReturn(goals);

        DailyGoalStreakDto result = getDailyGoalStreakUseCase.execute(1, today);

        assertThat(result.currentStreak()).isEqualTo(1);
        assertThat(result.longestStreak()).isEqualTo(3);
        assertThat(result.totalAchievedDays()).isEqualTo(4);
    }

    @Test
    @DisplayName("日付の空白がある場合ストリークが途切れる")
    void execute_dateGap_breaksStreak() {
        LocalDate today = LocalDate.now();
        List<DailyGoal> goals = List.of(
                goal(today, 3, 3),
                // today.minusDays(1) の記録なし → ストリーク切れ
                goal(today.minusDays(2), 3, 3),
                goal(today.minusDays(3), 3, 3)
        );

        when(dailyGoalRepository.findByUserIdOrderByGoalDateDesc(1)).thenReturn(goals);

        DailyGoalStreakDto result = getDailyGoalStreakUseCase.execute(1, today);

        assertThat(result.currentStreak()).isEqualTo(1);
        assertThat(result.longestStreak()).isEqualTo(2);
        assertThat(result.totalAchievedDays()).isEqualTo(3);
    }
}
