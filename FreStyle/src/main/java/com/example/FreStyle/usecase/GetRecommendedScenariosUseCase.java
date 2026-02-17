package com.example.FreStyle.usecase;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.RecommendedScenarioDto;
import com.example.FreStyle.dto.RecommendedScenarioDto.ScenarioRecommendation;
import com.example.FreStyle.entity.AiChatSession;
import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.repository.AiChatSessionRepository;
import com.example.FreStyle.repository.CommunicationScoreRepository;
import com.example.FreStyle.repository.PracticeScenarioRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetRecommendedScenariosUseCase {

    private final AiChatSessionRepository aiChatSessionRepository;
    private final CommunicationScoreRepository communicationScoreRepository;
    private final PracticeScenarioRepository practiceScenarioRepository;

    @Transactional(readOnly = true)
    public RecommendedScenarioDto execute(Integer userId) {
        List<AiChatSession> sessions = aiChatSessionRepository.findByUserIdOrderByCreatedAtDesc(userId);

        // セッションID → シナリオID のマッピング（練習セッションのみ）
        Map<Integer, Integer> sessionToScenario = sessions.stream()
                .filter(s -> s.getScenarioId() != null)
                .collect(Collectors.toMap(AiChatSession::getId, AiChatSession::getScenarioId));

        if (sessionToScenario.isEmpty()) {
            return new RecommendedScenarioDto(List.of());
        }

        // 全スコアを取得し、シナリオIDごとにグルーピング
        List<CommunicationScore> allScores = communicationScoreRepository.findByUserIdOrderByCreatedAtDesc(userId);

        Map<Integer, List<Integer>> scenarioScores = new HashMap<>();
        Map<Integer, java.util.Set<Integer>> scenarioSessions = new HashMap<>();

        for (CommunicationScore cs : allScores) {
            Integer sessionId = cs.getSession().getId();
            Integer scenarioId = sessionToScenario.get(sessionId);
            if (scenarioId != null) {
                scenarioScores.computeIfAbsent(scenarioId, k -> new ArrayList<>()).add(cs.getScore());
                scenarioSessions.computeIfAbsent(scenarioId, k -> new java.util.HashSet<>()).add(sessionId);
            }
        }

        // シナリオごとの平均スコアを計算し、低い順にソート
        List<ScenarioRecommendation> recommendations = scenarioScores.entrySet().stream()
                .map(entry -> {
                    Integer scenarioId = entry.getKey();
                    double avg = entry.getValue().stream().mapToInt(i -> i).average().orElse(0);
                    int practiceCount = scenarioSessions.get(scenarioId).size();
                    return practiceScenarioRepository.findById(scenarioId)
                            .map(ps -> new ScenarioRecommendation(
                                    ps.getId(), ps.getName(), ps.getCategory(),
                                    ps.getDifficulty(), avg, practiceCount))
                            .orElse(null);
                })
                .filter(r -> r != null)
                .sorted(Comparator.comparingDouble(ScenarioRecommendation::averageScore))
                .limit(5)
                .toList();

        return new RecommendedScenarioDto(recommendations);
    }
}
