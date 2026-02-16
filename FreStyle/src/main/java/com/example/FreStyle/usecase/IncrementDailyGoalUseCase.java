package com.example.FreStyle.usecase;

import java.time.LocalDate;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.DailyGoalDto;
import com.example.FreStyle.entity.DailyGoal;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.DailyGoalRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class IncrementDailyGoalUseCase {

    private final DailyGoalRepository dailyGoalRepository;

    @Transactional
    public DailyGoalDto execute(User user) {
        LocalDate today = LocalDate.now();
        DailyGoal goal = dailyGoalRepository.findByUserIdAndGoalDate(user.getId(), today)
                .orElseGet(() -> {
                    DailyGoal newGoal = new DailyGoal();
                    newGoal.setUser(user);
                    newGoal.setGoalDate(today);
                    newGoal.setTarget(3);
                    newGoal.setCompleted(0);
                    return newGoal;
                });
        goal.setCompleted(goal.getCompleted() + 1);
        dailyGoalRepository.save(goal);
        return new DailyGoalDto(goal.getGoalDate().toString(), goal.getTarget(), goal.getCompleted());
    }
}
