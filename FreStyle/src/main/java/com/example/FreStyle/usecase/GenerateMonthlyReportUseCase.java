package com.example.FreStyle.usecase;

import java.sql.Timestamp;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.LearningReportDto;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.entity.LearningReport;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.mapper.LearningReportMapper;
import com.example.FreStyle.repository.CommunicationScoreRepository;
import com.example.FreStyle.repository.LearningReportRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GenerateMonthlyReportUseCase {

    private final CommunicationScoreRepository communicationScoreRepository;
    private final LearningReportRepository learningReportRepository;
    private final LearningReportMapper learningReportMapper;

    @Transactional
    public LearningReportDto execute(User user, Integer year, Integer month) {
        List<CommunicationScore> allScores = communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(user.getId());

        List<CommunicationScore> monthlyScores = allScores.stream()
                .filter(s -> {
                    Timestamp ts = s.getCreatedAt();
                    if (ts == null) return false;
                    LocalDateTime ldt = ts.toLocalDateTime();
                    return ldt.getYear() == year && ldt.getMonthValue() == month;
                })
                .collect(Collectors.toList());

        long totalSessions = monthlyScores.stream()
                .map(s -> s.getSession().getId())
                .distinct()
                .count();

        double averageScore = monthlyScores.stream()
                .mapToInt(CommunicationScore::getScore)
                .average()
                .orElse(0.0);

        long practiceDays = monthlyScores.stream()
                .map(s -> s.getCreatedAt().toLocalDateTime().toLocalDate())
                .distinct()
                .count();

        String bestAxis = null;
        String worstAxis = null;
        if (!monthlyScores.isEmpty()) {
            Map<String, Double> axisMeans = monthlyScores.stream()
                    .collect(Collectors.groupingBy(CommunicationScore::getAxisName,
                            Collectors.averagingInt(CommunicationScore::getScore)));
            bestAxis = axisMeans.entrySet().stream()
                    .max(Map.Entry.comparingByValue())
                    .map(Map.Entry::getKey)
                    .orElse(null);
            worstAxis = axisMeans.entrySet().stream()
                    .min(Map.Entry.comparingByValue())
                    .map(Map.Entry::getKey)
                    .orElse(null);
            if (bestAxis != null && bestAxis.equals(worstAxis)) {
                worstAxis = null;
            }
        }

        // 前月のスコア取得
        int prevMonth = month == 1 ? 12 : month - 1;
        int prevYear = month == 1 ? year - 1 : year;
        Double previousAverageScore = learningReportRepository
                .findByUserIdAndYearAndMonth(user.getId(), prevYear, prevMonth)
                .map(LearningReport::getAverageScore)
                .orElse(null);

        LearningReport report = learningReportRepository
                .findByUserIdAndYearAndMonth(user.getId(), year, month)
                .orElseGet(() -> {
                    LearningReport newReport = new LearningReport();
                    newReport.setUser(user);
                    newReport.setYear(year);
                    newReport.setMonth(month);
                    return newReport;
                });

        report.setTotalSessions((int) totalSessions);
        report.setAverageScore(averageScore);
        report.setPreviousAverageScore(previousAverageScore);
        report.setBestAxis(bestAxis);
        report.setWorstAxis(worstAxis);
        report.setPracticeDays((int) practiceDays);

        LearningReport saved = learningReportRepository.save(report);
        return learningReportMapper.toDto(saved);
    }
}
