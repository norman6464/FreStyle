package com.example.FreStyle.usecase;

import static org.mockito.ArgumentMatchers.*;
import static org.mockito.Mockito.*;

import java.util.Optional;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.entity.ScoreGoal;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.ScoreGoalRepository;

@ExtendWith(MockitoExtension.class)
class SaveScoreGoalUseCaseTest {

    @Mock
    private ScoreGoalRepository scoreGoalRepository;

    @InjectMocks
    private SaveScoreGoalUseCase useCase;

    @Test
    @DisplayName("既存の目標スコアを更新する")
    void execute_updatesExistingGoal() {
        User user = new User();
        user.setId(1);
        ScoreGoal existing = new ScoreGoal();
        existing.setUser(user);
        existing.setGoalScore(7.0);
        when(scoreGoalRepository.findByUserId(1)).thenReturn(Optional.of(existing));

        useCase.execute(user, 8.5);

        verify(scoreGoalRepository).save(argThat(g -> g.getGoalScore() == 8.5));
    }

    @Test
    @DisplayName("新規の目標スコアを作成する")
    void execute_createsNewGoal() {
        User user = new User();
        user.setId(1);
        when(scoreGoalRepository.findByUserId(1)).thenReturn(Optional.empty());

        useCase.execute(user, 9.0);

        verify(scoreGoalRepository).save(argThat(g ->
                g.getUser().equals(user) && g.getGoalScore() == 9.0));
    }
}
