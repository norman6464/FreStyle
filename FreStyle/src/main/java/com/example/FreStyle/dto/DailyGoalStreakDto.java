package com.example.FreStyle.dto;

public record DailyGoalStreakDto(
        int currentStreak,
        int longestStreak,
        int totalAchievedDays) {
}
