package com.example.FreStyle.usecase;

import java.sql.Timestamp;
import java.time.LocalDateTime;
import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.ScoreTrendDto;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.repository.CommunicationScoreRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetScoreTrendUseCase {

    private final CommunicationScoreRepository communicationScoreRepository;

    @Transactional(readOnly = true)
    public ScoreTrendDto execute(Integer userId, int days) {
        LocalDateTime cutoff = LocalDateTime.now().minusDays(days);
        List<CommunicationScore> allScores = communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(userId);

        List<CommunicationScore> filteredScores = allScores.stream()
                .filter(s -> {
                    Timestamp ts = s.getCreatedAt();
                    return ts != null && ts.toLocalDateTime().isAfter(cutoff);
                })
                .collect(Collectors.toList());

        if (filteredScores.isEmpty()) {
            return new ScoreTrendDto(days, List.of(), 0.0, null, 0);
        }

        Map<Integer, List<CommunicationScore>> bySession = filteredScores.stream()
                .collect(Collectors.groupingBy(s -> s.getSession().getId()));

        List<ScoreTrendDto.SessionScore> sessionScores = bySession.entrySet().stream()
                .map(entry -> {
                    double avg = entry.getValue().stream()
                            .mapToInt(CommunicationScore::getScore)
                            .average()
                            .orElse(0.0);
                    Timestamp earliest = entry.getValue().stream()
                            .map(CommunicationScore::getCreatedAt)
                            .min(Comparator.naturalOrder())
                            .orElse(null);
                    String createdAt = earliest != null ? earliest.toLocalDateTime().toString() : null;
                    return new ScoreTrendDto.SessionScore(entry.getKey(), avg, createdAt);
                })
                .sorted(Comparator.comparing(ScoreTrendDto.SessionScore::createdAt,
                        Comparator.nullsLast(Comparator.naturalOrder())))
                .collect(Collectors.toList());

        double overallAverage = filteredScores.stream()
                .mapToInt(CommunicationScore::getScore)
                .average()
                .orElse(0.0);

        ScoreTrendDto.SessionScore bestSession = sessionScores.stream()
                .max(Comparator.comparingDouble(ScoreTrendDto.SessionScore::averageScore))
                .orElse(null);

        return new ScoreTrendDto(days, sessionScores, overallAverage, bestSession, sessionScores.size());
    }
}
