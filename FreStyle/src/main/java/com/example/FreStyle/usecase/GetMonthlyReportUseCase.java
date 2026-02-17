package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.LearningReportDto;
import com.example.FreStyle.mapper.LearningReportMapper;
import com.example.FreStyle.repository.LearningReportRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetMonthlyReportUseCase {

    private final LearningReportRepository learningReportRepository;
    private final LearningReportMapper learningReportMapper;

    @Transactional(readOnly = true)
    public LearningReportDto execute(Integer userId, Integer year, Integer month) {
        return learningReportRepository.findByUserIdAndYearAndMonth(userId, year, month)
                .map(learningReportMapper::toDto)
                .orElse(null);
    }
}
