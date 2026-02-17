package com.example.FreStyle.dto;

import java.util.List;

public record AxisAnalysisDto(
        List<AxisStat> axisStats,
        String bestAxis,
        String worstAxis,
        int totalEvaluations) {

    public record AxisStat(String axisName, double averageScore, int count) {
    }
}
