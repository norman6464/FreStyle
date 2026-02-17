package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.LearningReportDto;
import com.example.FreStyle.entity.LearningReport;
import com.example.FreStyle.repository.LearningReportRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetMonthlyReportUseCase {

    private final LearningReportRepository learningReportRepository;

    @Transactional(readOnly = true)
    public LearningReportDto execute(Integer userId, Integer year, Integer month) {
        return learningReportRepository.findByUserIdAndYearAndMonth(userId, year, month)
                .map(this::toDto)
                .orElse(null);
    }

    private LearningReportDto toDto(LearningReport report) {
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
