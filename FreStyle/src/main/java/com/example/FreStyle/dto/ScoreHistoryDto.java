package com.example.FreStyle.dto;

import java.sql.Timestamp;
import java.util.List;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class ScoreHistoryDto {

    private Integer sessionId;
    private String sessionTitle;
    private double overallScore;
    private List<ScoreCardDto.AxisScoreDto> scores;
    private Timestamp createdAt;
}
