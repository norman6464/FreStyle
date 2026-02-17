package com.example.FreStyle.usecase;

import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.AxisAnalysisDto;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.repository.CommunicationScoreRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetAxisAnalysisUseCase {

    private final CommunicationScoreRepository communicationScoreRepository;

    @Transactional(readOnly = true)
    public AxisAnalysisDto execute(Integer userId) {
        List<CommunicationScore> scores = communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(userId);

        if (scores.isEmpty()) {
            return new AxisAnalysisDto(List.of(), null, null, 0);
        }

        Map<String, List<CommunicationScore>> grouped = scores.stream()
                .collect(Collectors.groupingBy(CommunicationScore::getAxisName));

        List<AxisAnalysisDto.AxisStat> axisStats = grouped.entrySet().stream()
                .map(entry -> {
                    double avg = entry.getValue().stream()
                            .mapToInt(CommunicationScore::getScore)
                            .average()
                            .orElse(0.0);
                    return new AxisAnalysisDto.AxisStat(entry.getKey(), avg, entry.getValue().size());
                })
                .sorted(Comparator.comparingDouble(AxisAnalysisDto.AxisStat::averageScore).reversed()
                        .thenComparing(AxisAnalysisDto.AxisStat::axisName))
                .toList();

        String bestAxis = axisStats.getFirst().axisName();
        String worstAxis = axisStats.getLast().axisName();

        return new AxisAnalysisDto(axisStats, bestAxis, worstAxis, scores.size());
    }
}
