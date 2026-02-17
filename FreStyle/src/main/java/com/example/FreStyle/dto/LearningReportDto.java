package com.example.FreStyle.dto;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class LearningReportDto {
    private Integer id;
    private Integer year;
    private Integer month;
    private Integer totalSessions;
    private Double averageScore;
    private Double previousAverageScore;
    private Double scoreChange;
    private String bestAxis;
    private String worstAxis;
    private Integer practiceDays;
    private String createdAt;
}
