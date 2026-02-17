package com.example.FreStyle.dto;

public record UserStatsDto(
        long totalSessions,
        long practiceSessionCount,
        long followerCount,
        long followingCount,
        double averageScore) {
}
