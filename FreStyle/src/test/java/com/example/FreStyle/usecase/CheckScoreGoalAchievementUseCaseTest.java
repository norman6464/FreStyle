package com.example.FreStyle.usecase;

import static org.mockito.Mockito.*;

import java.util.List;
import java.util.Optional;

import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.entity.CommunicationScore;
import com.example.FreStyle.entity.ScoreGoal;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.repository.CommunicationScoreRepository;
import com.example.FreStyle.repository.ScoreGoalRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("CheckScoreGoalAchievementUseCase テスト")
class CheckScoreGoalAchievementUseCaseTest {

    @Mock
    private ScoreGoalRepository scoreGoalRepository;

    @Mock
    private CommunicationScoreRepository communicationScoreRepository;

    @Mock
    private CreateNotificationUseCase createNotificationUseCase;

    @InjectMocks
    private CheckScoreGoalAchievementUseCase checkScoreGoalAchievementUseCase;

    private User testUser;

    @BeforeEach
    void setUp() {
        testUser = new User();
        testUser.setId(1);
        testUser.setName("テストユーザー");
    }

    @Test
    @DisplayName("目標達成時に通知が作成される")
    void execute_goalAchieved_createsNotification() {
        ScoreGoal goal = new ScoreGoal();
        goal.setGoalScore(70.0);
        when(scoreGoalRepository.findByUserId(1)).thenReturn(Optional.of(goal));

        CommunicationScore s1 = new CommunicationScore();
        s1.setScore(80);
        CommunicationScore s2 = new CommunicationScore();
        s2.setScore(75);
        when(communicationScoreRepository.findBySessionId(100)).thenReturn(List.of(s1, s2));

        checkScoreGoalAchievementUseCase.execute(testUser, 100);

        verify(createNotificationUseCase).execute(
                eq(testUser), eq("SCORE_GOAL_ACHIEVED"),
                anyString(), anyString(), eq(100));
    }

    @Test
    @DisplayName("目標未達成時に通知が作成されない")
    void execute_goalNotAchieved_noNotification() {
        ScoreGoal goal = new ScoreGoal();
        goal.setGoalScore(90.0);
        when(scoreGoalRepository.findByUserId(1)).thenReturn(Optional.of(goal));

        CommunicationScore s1 = new CommunicationScore();
        s1.setScore(60);
        CommunicationScore s2 = new CommunicationScore();
        s2.setScore(70);
        when(communicationScoreRepository.findBySessionId(100)).thenReturn(List.of(s1, s2));

        checkScoreGoalAchievementUseCase.execute(testUser, 100);

        verify(createNotificationUseCase, never()).execute(any(), anyString(), anyString(), anyString(), any());
    }

    @Test
    @DisplayName("スコア目標が未設定の場合何もしない")
    void execute_noGoal_doesNothing() {
        when(scoreGoalRepository.findByUserId(1)).thenReturn(Optional.empty());

        checkScoreGoalAchievementUseCase.execute(testUser, 100);

        verify(communicationScoreRepository, never()).findBySessionId(anyInt());
        verify(createNotificationUseCase, never()).execute(any(), anyString(), anyString(), anyString(), any());
    }

    @Test
    @DisplayName("ちょうど目標スコアの場合に通知が作成される")
    void execute_exactGoalScore_createsNotification() {
        ScoreGoal goal = new ScoreGoal();
        goal.setGoalScore(75.0);
        when(scoreGoalRepository.findByUserId(1)).thenReturn(Optional.of(goal));

        CommunicationScore s1 = new CommunicationScore();
        s1.setScore(75);
        when(communicationScoreRepository.findBySessionId(100)).thenReturn(List.of(s1));

        checkScoreGoalAchievementUseCase.execute(testUser, 100);

        verify(createNotificationUseCase).execute(
                eq(testUser), eq("SCORE_GOAL_ACHIEVED"),
                anyString(), anyString(), eq(100));
    }

    @Test
    @DisplayName("セッションにスコアがない場合通知が作成されない")
    void execute_noScores_noNotification() {
        ScoreGoal goal = new ScoreGoal();
        goal.setGoalScore(70.0);
        when(scoreGoalRepository.findByUserId(1)).thenReturn(Optional.of(goal));
        when(communicationScoreRepository.findBySessionId(100)).thenReturn(List.of());

        checkScoreGoalAchievementUseCase.execute(testUser, 100);

        verify(createNotificationUseCase, never()).execute(any(), anyString(), anyString(), anyString(), any());
    }
}
