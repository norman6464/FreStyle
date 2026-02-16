package com.example.FreStyle.usecase;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import java.time.LocalDate;
import java.util.Optional;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.DailyGoalDto;
import com.example.FreStyle.entity.DailyGoal;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.DailyGoalRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("IncrementDailyGoalUseCase テスト")
class IncrementDailyGoalUseCaseTest {

    @Mock
    private DailyGoalRepository dailyGoalRepository;

    @InjectMocks
    private IncrementDailyGoalUseCase incrementDailyGoalUseCase;

    @Test
    @DisplayName("既存ゴールの完了数をインクリメントする")
    void execute_existingGoal_incrementsCompleted() {
        User user = new User();
        user.setId(1);
        DailyGoal goal = new DailyGoal();
        goal.setUser(user);
        goal.setGoalDate(LocalDate.now());
        goal.setTarget(5);
        goal.setCompleted(2);
        when(dailyGoalRepository.findByUserIdAndGoalDate(1, LocalDate.now()))
                .thenReturn(Optional.of(goal));

        DailyGoalDto result = incrementDailyGoalUseCase.execute(user);

        assertEquals(3, result.getCompleted());
        verify(dailyGoalRepository).save(argThat(g -> g.getCompleted() == 3));
    }

    @Test
    @DisplayName("ゴールが存在しない場合は新規作成してインクリメントする")
    void execute_noGoal_createsAndIncrements() {
        User user = new User();
        user.setId(1);
        when(dailyGoalRepository.findByUserIdAndGoalDate(1, LocalDate.now()))
                .thenReturn(Optional.empty());

        DailyGoalDto result = incrementDailyGoalUseCase.execute(user);

        assertEquals(1, result.getCompleted());
        assertEquals(3, result.getTarget());
        verify(dailyGoalRepository).save(argThat(g ->
                g.getCompleted() == 1 && g.getTarget() == 3));
    }

    @Test
    @DisplayName("レスポンスの日付が今日になる")
    void execute_returnsToday() {
        User user = new User();
        user.setId(1);
        DailyGoal goal = new DailyGoal();
        goal.setUser(user);
        goal.setGoalDate(LocalDate.now());
        goal.setTarget(3);
        goal.setCompleted(0);
        when(dailyGoalRepository.findByUserIdAndGoalDate(1, LocalDate.now()))
                .thenReturn(Optional.of(goal));

        DailyGoalDto result = incrementDailyGoalUseCase.execute(user);

        assertEquals(LocalDate.now().toString(), result.getDate());
    }

    @Test
    @DisplayName("目標達成時もインクリメントされる")
    void execute_atTarget_incrementsBeyond() {
        User user = new User();
        user.setId(1);
        DailyGoal goal = new DailyGoal();
        goal.setUser(user);
        goal.setGoalDate(LocalDate.now());
        goal.setTarget(5);
        goal.setCompleted(4);
        when(dailyGoalRepository.findByUserIdAndGoalDate(1, LocalDate.now()))
                .thenReturn(Optional.of(goal));

        DailyGoalDto result = incrementDailyGoalUseCase.execute(user);

        assertEquals(5, result.getCompleted());
        assertEquals(5, result.getTarget());
    }
}
