package com.example.FreStyle.usecase;

import java.time.LocalDate;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.DailyGoal;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.DailyGoalRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class SetDailyGoalTargetUseCase {

    private final DailyGoalRepository dailyGoalRepository;

    @Transactional
    public void execute(User user, Integer target) {
        LocalDate today = LocalDate.now();
        DailyGoal goal = dailyGoalRepository.findByUserIdAndGoalDate(user.getId(), today)
                .orElseGet(() -> {
                    DailyGoal newGoal = new DailyGoal();
                    newGoal.setUser(user);
                    newGoal.setGoalDate(today);
                    newGoal.setCompleted(0);
                    return newGoal;
                });
        goal.setTarget(target);
        dailyGoalRepository.save(goal);
    }
}
