package com.example.FreStyle.usecase;

import static org.mockito.Mockito.*;

import java.time.LocalDate;
import java.util.Optional;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.entity.DailyGoal;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.DailyGoalRepository;
import com.example.FreStyle.service.UserIdentityService;

@ExtendWith(MockitoExtension.class)
@DisplayName("SetDailyGoalTargetUseCase テスト")
class SetDailyGoalTargetUseCaseTest {

    @Mock
    private DailyGoalRepository dailyGoalRepository;

    @Mock
    private UserIdentityService userIdentityService;

    @InjectMocks
    private SetDailyGoalTargetUseCase setDailyGoalTargetUseCase;

    @Test
    @DisplayName("既存ゴールがある場合はターゲットを更新する")
    void execute_existingGoal_updatesTarget() {
        User user = new User();
        user.setId(1);
        DailyGoal goal = new DailyGoal();
        goal.setUser(user);
        goal.setGoalDate(LocalDate.now());
        goal.setTarget(3);
        goal.setCompleted(1);
        when(dailyGoalRepository.findByUserIdAndGoalDate(1, LocalDate.now()))
                .thenReturn(Optional.of(goal));

        setDailyGoalTargetUseCase.execute(user, 5);

        verify(dailyGoalRepository).save(argThat(g -> g.getTarget() == 5));
    }

    @Test
    @DisplayName("ゴールが存在しない場合は新規作成する")
    void execute_noGoal_createsNew() {
        User user = new User();
        user.setId(1);
        when(dailyGoalRepository.findByUserIdAndGoalDate(1, LocalDate.now()))
                .thenReturn(Optional.empty());

        setDailyGoalTargetUseCase.execute(user, 7);

        verify(dailyGoalRepository).save(argThat(g ->
                g.getTarget() == 7 && g.getCompleted() == 0 && g.getUser().getId() == 1));
    }
}
