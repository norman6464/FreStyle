package com.example.FreStyle.service;

import java.sql.Timestamp;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.RankingDto;
import com.example.FreStyle.dto.RankingDto.RankingEntryDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.CommunicationScoreRepository;
import com.example.FreStyle.repository.UserRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class RankingService {

    private final CommunicationScoreRepository communicationScoreRepository;
    private final UserRepository userRepository;

    @Transactional(readOnly = true)
    public RankingDto getRanking(String period, Integer currentUserId) {
        Timestamp cutoff = calculateCutoff(period);
        List<Object[]> results = communicationScoreRepository.findUserAverageScoresAfter(cutoff);

        List<RankingEntryDto> entries = new ArrayList<>();
        RankingEntryDto myRanking = null;

        for (int i = 0; i < results.size(); i++) {
            Object[] row = results.get(i);
            Integer userId = (Integer) row[0];
            double avgScore = ((Number) row[1]).doubleValue();
            int sessionCount = ((Number) row[2]).intValue();

            User user = userRepository.findById(userId).orElse(null);
            if (user == null) continue;

            RankingEntryDto entry = new RankingEntryDto(
                i + 1,
                userId,
                user.getName(),
                user.getIconUrl(),
                Math.round(avgScore * 10.0) / 10.0,
                sessionCount
            );
            entries.add(entry);

            if (userId.equals(currentUserId)) {
                myRanking = entry;
            }
        }

        return new RankingDto(entries, myRanking);
    }

    private Timestamp calculateCutoff(String period) {
        LocalDateTime now = LocalDateTime.now();
        LocalDateTime cutoff = switch (period) {
            case "weekly" -> now.minusWeeks(1);
            case "monthly" -> now.minusMonths(1);
            default -> now.minusMonths(1);
        };
        return Timestamp.valueOf(cutoff);
    }
}
