package com.example.FreStyle.dto;

import java.util.List;

public record CommunicationScoreSummaryDto(
        int totalSessions,
        double overallAverage,
        List<AxisAverage> axisAverages,
        String bestAxis,
        String worstAxis) {

    public record AxisAverage(String axisName, double average, long count) {
    }
}
