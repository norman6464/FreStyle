package com.example.FreStyle.usecase;

import java.util.List;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.LearningReportDto;
import com.example.FreStyle.entity.LearningReport;
import com.example.FreStyle.repository.LearningReportRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetReportListUseCase {

    private final LearningReportRepository learningReportRepository;

    @Transactional(readOnly = true)
    public List<LearningReportDto> execute(Integer userId) {
        return learningReportRepository.findByUserIdOrderByYearDescMonthDesc(userId)
                .stream()
                .map(this::toDto)
                .collect(Collectors.toList());
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
