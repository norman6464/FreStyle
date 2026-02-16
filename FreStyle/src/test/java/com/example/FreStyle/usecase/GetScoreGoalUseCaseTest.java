package com.example.FreStyle.usecase;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import java.util.Optional;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.ScoreGoalDto;
import com.example.FreStyle.entity.ScoreGoal;
import com.example.FreStyle.repository.ScoreGoalRepository;

@ExtendWith(MockitoExtension.class)
class GetScoreGoalUseCaseTest {

    @Mock
    private ScoreGoalRepository scoreGoalRepository;

    @InjectMocks
    private GetScoreGoalUseCase useCase;

    @Test
    @DisplayName("ユーザーの目標スコアを返す")
    void execute_returnsGoal() {
        ScoreGoal goal = new ScoreGoal();
        goal.setGoalScore(8.5);
        when(scoreGoalRepository.findByUserId(1)).thenReturn(Optional.of(goal));

        ScoreGoalDto result = useCase.execute(1);

        assertNotNull(result);
        assertEquals(8.5, result.getGoalScore());
    }

    @Test
    @DisplayName("目標スコアが未設定の場合はnullを返す")
    void execute_returnsNullWhenNotFound() {
        when(scoreGoalRepository.findByUserId(999)).thenReturn(Optional.empty());

        ScoreGoalDto result = useCase.execute(999);

        assertNull(result);
    }
}
