package com.example.FreStyle.dto;

import java.util.List;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class ScoreCardDto {

    private Integer sessionId;
    private List<AxisScoreDto> scores;
    private double overallScore;

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class AxisScoreDto {
        private String axis;
        private int score;
        private String comment;
    }
}
