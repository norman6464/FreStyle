package com.example.FreStyle.dto;

import java.util.List;

public record ScoreCardDto(
        Integer sessionId,
        List<AxisScoreDto> scores,
        double overallScore) {

    public record AxisScoreDto(
            String axis,
            int score,
            String comment) {
    }
}
