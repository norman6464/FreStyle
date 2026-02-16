package com.example.FreStyle.usecase;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.ScoreGoalDto;
import com.example.FreStyle.repository.ScoreGoalRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetScoreGoalUseCase {

    private final ScoreGoalRepository scoreGoalRepository;

    @Transactional(readOnly = true)
    public ScoreGoalDto execute(Integer userId) {
        return scoreGoalRepository.findByUserId(userId)
                .map(goal -> new ScoreGoalDto(goal.getGoalScore()))
                .orElse(null);
    }
}
