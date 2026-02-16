package com.example.FreStyle.usecase;

import java.time.LocalDate;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.DailyGoalDto;
import com.example.FreStyle.repository.DailyGoalRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetTodayDailyGoalUseCase {

    private final DailyGoalRepository dailyGoalRepository;

    @Transactional(readOnly = true)
    public DailyGoalDto execute(Integer userId) {
        LocalDate today = LocalDate.now();
        return dailyGoalRepository.findByUserIdAndGoalDate(userId, today)
                .map(goal -> new DailyGoalDto(goal.getGoalDate().toString(), goal.getTarget(), goal.getCompleted()))
                .orElse(new DailyGoalDto(today.toString(), 3, 0));
    }
}
