package com.example.FreStyle.usecase;

import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.LearningReportDto;
import com.example.FreStyle.mapper.LearningReportMapper;
import com.example.FreStyle.repository.LearningReportRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetReportListUseCase {

    private final LearningReportRepository learningReportRepository;
    private final LearningReportMapper learningReportMapper;

    @Transactional(readOnly = true)
    public List<LearningReportDto> execute(Integer userId) {
        return learningReportRepository.findByUserIdOrderByYearDescMonthDesc(userId)
                .stream()
                .map(learningReportMapper::toDto)
                .toList();
    }
}
