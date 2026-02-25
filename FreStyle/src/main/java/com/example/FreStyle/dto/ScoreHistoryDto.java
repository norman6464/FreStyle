package com.example.FreStyle.dto;

import java.sql.Timestamp;
import java.util.List;

public record ScoreHistoryDto(
        Integer sessionId,
        String sessionTitle,
        Integer scenarioId,
        double overallScore,
        List<ScoreCardDto.AxisScoreDto> scores,
        Timestamp createdAt) {
}
