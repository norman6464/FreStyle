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
@DisplayName("GetTodayDailyGoalUseCase テスト")
class GetTodayDailyGoalUseCaseTest {

    @Mock
    private DailyGoalRepository dailyGoalRepository;

    @InjectMocks
    private GetTodayDailyGoalUseCase getTodayDailyGoalUseCase;

    @Test
    @DisplayName("今日のゴールが存在する場合はそれを返す")
    void execute_existingGoal_returnsDto() {
        User user = new User();
        user.setId(1);
        DailyGoal goal = new DailyGoal();
        goal.setId(10);
        goal.setUser(user);
        goal.setGoalDate(LocalDate.now());
        goal.setTarget(5);
        goal.setCompleted(2);
        when(dailyGoalRepository.findByUserIdAndGoalDate(1, LocalDate.now()))
                .thenReturn(Optional.of(goal));

        DailyGoalDto result = getTodayDailyGoalUseCase.execute(1);

        assertEquals(LocalDate.now().toString(), result.date());
        assertEquals(5, result.target());
        assertEquals(2, result.completed());
    }

    @Test
    @DisplayName("今日のゴールが存在しない場合はデフォルト値を返す")
    void execute_noGoal_returnsDefault() {
        when(dailyGoalRepository.findByUserIdAndGoalDate(1, LocalDate.now()))
                .thenReturn(Optional.empty());

        DailyGoalDto result = getTodayDailyGoalUseCase.execute(1);

        assertEquals(LocalDate.now().toString(), result.date());
        assertEquals(3, result.target());
        assertEquals(0, result.completed());
    }
}
