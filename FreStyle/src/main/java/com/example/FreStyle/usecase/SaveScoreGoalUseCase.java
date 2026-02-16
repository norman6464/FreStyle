package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.entity.ScoreGoal;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.ScoreGoalRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class SaveScoreGoalUseCase {

    private final ScoreGoalRepository scoreGoalRepository;

    @Transactional
    public void execute(User user, Double goalScore) {
        ScoreGoal scoreGoal = scoreGoalRepository.findByUserId(user.getId())
                .orElseGet(() -> {
                    ScoreGoal newGoal = new ScoreGoal();
                    newGoal.setUser(user);
                    return newGoal;
                });
        scoreGoal.setGoalScore(goalScore);
        scoreGoalRepository.save(scoreGoal);
    }
}
