package com.example.FreStyle.usecase;

import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.AiChatSessionStatsDto;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.repository.AiChatSessionRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetAiChatSessionStatsUseCase {

    private final AiChatSessionRepository aiChatSessionRepository;

    @Transactional(readOnly = true)
    public AiChatSessionStatsDto execute(Integer userId) {
        List<AiChatSession> sessions = aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(userId);

        if (sessions.isEmpty()) {
            return new AiChatSessionStatsDto(0, List.of(), List.of());
        }

        List<AiChatSessionStatsDto.TypeCount> typeCounts = groupAndSort(
                sessions.stream()
                        .map(AiChatSession::getSessionType)
                        .filter(Objects::nonNull)
                        .collect(Collectors.groupingBy(t -> t, Collectors.counting())),
                AiChatSessionStatsDto.TypeCount::new);

        List<AiChatSessionStatsDto.SceneCount> sceneCounts = groupAndSort(
                sessions.stream()
                        .map(AiChatSession::getScene)
                        .filter(Objects::nonNull)
                        .collect(Collectors.groupingBy(s -> s, Collectors.counting())),
                AiChatSessionStatsDto.SceneCount::new);

        return new AiChatSessionStatsDto(sessions.size(), typeCounts, sceneCounts);
    }

    private <T> List<T> groupAndSort(Map<String, Long> countMap,
            java.util.function.BiFunction<String, Long, T> constructor) {
        return countMap.entrySet().stream()
                .sorted(Map.Entry.<String, Long>comparingByValue(Comparator.reverseOrder())
                        .thenComparing(Map.Entry.comparingByKey()))
                .map(e -> constructor.apply(e.getKey(), e.getValue()))
                .toList();
    }
}
