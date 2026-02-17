package com.example.FreStyle.dto;

import java.util.List;

public record ScoreTrendDto(
        int days,
        List<SessionScore> sessionScores,
        double overallAverage,
        SessionScore bestSession,
        int totalSessions,
        Double improvement) {

    public record SessionScore(
            Integer sessionId,
            double averageScore,
            String createdAt) {
    }
}
