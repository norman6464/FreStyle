package com.example.FreStyle.usecase;

import java.util.List;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.CommunicationScoreRepository;
import com.example.FreStyle.repository.ScoreGoalRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class CheckScoreGoalAchievementUseCase {

    private final ScoreGoalRepository scoreGoalRepository;
    private final CommunicationScoreRepository communicationScoreRepository;
    private final CreateNotificationUseCase createNotificationUseCase;

    @Transactional
    public void execute(User user, Integer sessionId) {
        scoreGoalRepository.findByUserId(user.getId()).ifPresent(goal -> {
            List<CommunicationScore> scores = communicationScoreRepository.findBySessionId(sessionId);
            if (scores.isEmpty()) {
                return;
            }

            double average = scores.stream()
                    .mapToInt(CommunicationScore::getScore)
                    .average()
                    .getAsDouble();

            if (average >= goal.getGoalScore()) {
                createNotificationUseCase.execute(
                        user,
                        "SCORE_GOAL_ACHIEVED",
                        "スコア目標達成！",
                        String.format("セッションの平均スコア %.1f が目標スコア %.1f を達成しました！", average, goal.getGoalScore()),
                        sessionId);
            }
        });
    }
}
