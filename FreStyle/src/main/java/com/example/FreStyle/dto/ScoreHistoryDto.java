package com.example.FreStyle.dto;

import java.sql.Timestamp;
import java.util.List;

public record ScoreHistoryDto(
        Integer sessionId,
        String sessionTitle,
        double overallScore,
        List<ScoreCardDto.AxisScoreDto> scores,
        Timestamp createdAt) {
}
