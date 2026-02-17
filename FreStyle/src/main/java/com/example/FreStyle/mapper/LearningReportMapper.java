package com.example.FreStyle.mapper;

import org.springframework.stereotype.Component;

import com.example.FreStyle.dto.LearningReportDto;
import com.example.FreStyle.entity.LearningReport;

@Component
public class LearningReportMapper {

    public LearningReportDto toDto(LearningReport report) {
        Double scoreChange = null;
        if (report.getPreviousAverageScore() != null) {
            scoreChange = report.getAverageScore() - report.getPreviousAverageScore();
        }
        return new LearningReportDto(
                report.getId(),
                report.getYear(),
                report.getMonth(),
                report.getTotalSessions(),
                report.getAverageScore(),
                report.getPreviousAverageScore(),
                scoreChange,
                report.getBestAxis(),
                report.getWorstAxis(),
                report.getPracticeDays(),
                report.getCreatedAt() != null ? report.getCreatedAt().toString() : null
        );
    }
}
