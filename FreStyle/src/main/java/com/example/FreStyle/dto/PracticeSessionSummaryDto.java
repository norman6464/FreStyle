package com.example.FreStyle.dto;

import java.sql.Timestamp;
import java.util.List;

public record PracticeSessionSummaryDto(
        Integer sessionId,
        String title,
        String sessionType,
        String scene,
        Timestamp createdAt,
        Long messageCount,
        List<ScoreDetail> scores,
        Double averageScore,
        String bestAxis,
        String worstAxis,
        String note,
        String scenarioName
) {
    public record ScoreDetail(
            String axisName,
            Integer score,
            String comment
    ) {}
}
