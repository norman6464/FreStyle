package com.example.FreStyle.dto;

public record LearningReportDto(
        Integer id,
        Integer year,
        Integer month,
        Integer totalSessions,
        Double averageScore,
        Double previousAverageScore,
        Double scoreChange,
        String bestAxis,
        String worstAxis,
        Integer practiceDays,
        String createdAt
) {}
