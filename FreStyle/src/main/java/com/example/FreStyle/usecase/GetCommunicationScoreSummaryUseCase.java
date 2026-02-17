package com.example.FreStyle.usecase;

import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.CommunicationScoreSummaryDto;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.repository.CommunicationScoreRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetCommunicationScoreSummaryUseCase {

    private final CommunicationScoreRepository communicationScoreRepository;

    @Transactional(readOnly = true)
    public CommunicationScoreSummaryDto execute(Integer userId) {
        List<CommunicationScore> scores = communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(userId);

        if (scores.isEmpty()) {
            return new CommunicationScoreSummaryDto(0, 0.0, List.of(), null, null);
        }

        int totalSessions = (int) scores.stream()
                .map(s -> s.getSession().getId())
                .distinct()
                .count();

        double overallAverage = scores.stream()
                .mapToInt(CommunicationScore::getScore)
                .average()
                .orElse(0.0);

        Map<String, List<CommunicationScore>> byAxis = scores.stream()
                .collect(Collectors.groupingBy(CommunicationScore::getAxisName));

        List<CommunicationScoreSummaryDto.AxisAverage> axisAverages = byAxis.entrySet().stream()
                .map(e -> new CommunicationScoreSummaryDto.AxisAverage(
                        e.getKey(),
                        e.getValue().stream().mapToInt(CommunicationScore::getScore).average().orElse(0.0),
                        e.getValue().size()))
                .sorted(Comparator.comparingDouble(CommunicationScoreSummaryDto.AxisAverage::average).reversed()
                        .thenComparing(CommunicationScoreSummaryDto.AxisAverage::axisName))
                .toList();

        String bestAxis = axisAverages.getFirst().axisName();
        String worstAxis = axisAverages.getLast().axisName();

        return new CommunicationScoreSummaryDto(totalSessions, overallAverage, axisAverages, bestAxis, worstAxis);
    }
}
