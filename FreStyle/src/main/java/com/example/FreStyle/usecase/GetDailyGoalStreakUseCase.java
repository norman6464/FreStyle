package com.example.FreStyle.usecase;

import java.time.LocalDate;
import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.DailyGoalStreakDto;
import com.example.FreStyle.entity.DailyGoal;
import com.example.FreStyle.repository.DailyGoalRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetDailyGoalStreakUseCase {

    private final DailyGoalRepository dailyGoalRepository;

    @Transactional(readOnly = true)
    public DailyGoalStreakDto execute(Integer userId, LocalDate today) {
        List<DailyGoal> goals = dailyGoalRepository.findByUserIdOrderByGoalDateDesc(userId);

        if (goals.isEmpty()) {
            return new DailyGoalStreakDto(0, 0, 0);
        }

        int totalAchievedDays = 0;
        int currentStreak = 0;
        int longestStreak = 0;
        int tempStreak = 0;
        LocalDate expectedDate = null;
        boolean currentStreakDetermined = false;

        for (DailyGoal goal : goals) {
            boolean achieved = goal.getCompleted() >= goal.getTarget();

            if (achieved) {
                totalAchievedDays++;
            }

            if (expectedDate == null) {
                // 最初のレコード: 今日 or 昨日でないとcurrentStreakにはならない
                if (achieved && (goal.getGoalDate().equals(today) || goal.getGoalDate().equals(today.minusDays(1)))) {
                    currentStreak = 1;
                    tempStreak = 1;
                    expectedDate = goal.getGoalDate().minusDays(1);
                } else if (achieved) {
                    // 今日/昨日ではないが達成 → currentStreakは0、tempStreakは開始
                    currentStreakDetermined = true;
                    tempStreak = 1;
                    expectedDate = goal.getGoalDate().minusDays(1);
                } else {
                    // 未達成の場合
                    if (goal.getGoalDate().equals(today)) {
                        // 今日が未達成なら昨日からのストリークを探す
                        expectedDate = today.minusDays(1);
                    } else {
                        currentStreakDetermined = true;
                        expectedDate = goal.getGoalDate().minusDays(1);
                    }
                }
            } else if (goal.getGoalDate().equals(expectedDate)) {
                if (achieved) {
                    tempStreak++;
                    if (!currentStreakDetermined) {
                        currentStreak = tempStreak;
                    }
                    expectedDate = expectedDate.minusDays(1);
                } else {
                    if (!currentStreakDetermined) {
                        currentStreakDetermined = true;
                    }
                    longestStreak = Math.max(longestStreak, tempStreak);
                    tempStreak = 0;
                    expectedDate = goal.getGoalDate().minusDays(1);
                }
            } else {
                // 日付に空白がある
                if (!currentStreakDetermined) {
                    currentStreakDetermined = true;
                }
                longestStreak = Math.max(longestStreak, tempStreak);
                if (achieved) {
                    tempStreak = 1;
                } else {
                    tempStreak = 0;
                }
                expectedDate = goal.getGoalDate().minusDays(1);
            }
        }

        longestStreak = Math.max(longestStreak, tempStreak);

        return new DailyGoalStreakDto(currentStreak, longestStreak, totalAchievedDays);
    }
}
