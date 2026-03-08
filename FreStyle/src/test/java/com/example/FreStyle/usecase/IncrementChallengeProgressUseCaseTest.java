package com.example.FreStyle.usecase;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

import java.sql.Date;
import java.util.Optional;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import com.example.FreStyle.dto.WeeklyChallengeDto;
import com.example.FreStyle.entity.User;
import com.example.FreStyle.entity.UserChallengeProgress;
import com.example.FreStyle.entity.WeeklyChallenge;
import com.example.FreStyle.repository.UserChallengeProgressRepository;
import com.example.FreStyle.repository.WeeklyChallengeRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("IncrementChallengeProgressUseCase テスト")
class IncrementChallengeProgressUseCaseTest {

    @Mock
    private WeeklyChallengeRepository weeklyChallengeRepository;

    @Mock
    private UserChallengeProgressRepository userChallengeProgressRepository;

    @InjectMocks
    private IncrementChallengeProgressUseCase incrementChallengeProgressUseCase;

    @Test
    @DisplayName("既存の進捗がある場合、セッション数をインクリメントする")
    void execute_existingProgress_incrementsCompletedSessions() {
        WeeklyChallenge challenge = new WeeklyChallenge();
        challenge.setId(1);
        challenge.setTitle("チャレンジ");
        challenge.setDescription("説明");
        challenge.setCategory("communication");
        challenge.setTargetSessions(3);
        challenge.setWeekStart(Date.valueOf("2026-03-02"));
        challenge.setWeekEnd(Date.valueOf("2026-03-08"));

        when(weeklyChallengeRepository.findByWeekStartLessThanEqualAndWeekEndGreaterThanEqual(
                any(Date.class), any(Date.class)))
                .thenReturn(Optional.of(challenge));

        UserChallengeProgress progress = new UserChallengeProgress();
        progress.setCompletedSessions(1);
        progress.setIsCompleted(false);
        progress.setChallenge(challenge);
        when(userChallengeProgressRepository.findByUserIdAndChallengeId(1, 1))
                .thenReturn(Optional.of(progress));

        WeeklyChallengeDto result = incrementChallengeProgressUseCase.execute(1);

        assertEquals(2, result.completedSessions());
        assertFalse(result.isCompleted());
        verify(userChallengeProgressRepository).save(argThat(p -> p.getCompletedSessions() == 2));
    }

    @Test
    @DisplayName("進捗が存在しない場合、新規作成してインクリメントする")
    void execute_noProgress_createsNewAndIncrements() {
        WeeklyChallenge challenge = new WeeklyChallenge();
        challenge.setId(1);
        challenge.setTitle("チャレンジ");
        challenge.setDescription("説明");
        challenge.setCategory("communication");
        challenge.setTargetSessions(3);
        challenge.setWeekStart(Date.valueOf("2026-03-02"));
        challenge.setWeekEnd(Date.valueOf("2026-03-08"));

        when(weeklyChallengeRepository.findByWeekStartLessThanEqualAndWeekEndGreaterThanEqual(
                any(Date.class), any(Date.class)))
                .thenReturn(Optional.of(challenge));

        when(userChallengeProgressRepository.findByUserIdAndChallengeId(1, 1))
                .thenReturn(Optional.empty());

        WeeklyChallengeDto result = incrementChallengeProgressUseCase.execute(1);

        assertEquals(1, result.completedSessions());
        assertFalse(result.isCompleted());
        verify(userChallengeProgressRepository).save(argThat(p ->
                p.getCompletedSessions() == 1 && !p.getIsCompleted()));
    }
}
