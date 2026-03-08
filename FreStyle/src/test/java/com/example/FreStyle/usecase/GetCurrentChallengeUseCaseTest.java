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
import com.example.FreStyle.entity.UserChallengeProgress;
import com.example.FreStyle.entity.WeeklyChallenge;
import com.example.FreStyle.repository.UserChallengeProgressRepository;
import com.example.FreStyle.repository.WeeklyChallengeRepository;

@ExtendWith(MockitoExtension.class)
@DisplayName("GetCurrentChallengeUseCase テスト")
class GetCurrentChallengeUseCaseTest {

    @Mock
    private WeeklyChallengeRepository weeklyChallengeRepository;

    @Mock
    private UserChallengeProgressRepository userChallengeProgressRepository;

    @InjectMocks
    private GetCurrentChallengeUseCase getCurrentChallengeUseCase;

    @Test
    @DisplayName("今週のチャレンジが存在する場合、ユーザー進捗と合わせて返す")
    void execute_withExistingChallenge_returnsChallengeWithProgress() {
        WeeklyChallenge challenge = new WeeklyChallenge();
        challenge.setId(1);
        challenge.setTitle("コミュニケーション強化週間");
        challenge.setDescription("今週は3回練習しよう");
        challenge.setCategory("communication");
        challenge.setTargetSessions(3);
        challenge.setWeekStart(Date.valueOf("2026-03-02"));
        challenge.setWeekEnd(Date.valueOf("2026-03-08"));

        when(weeklyChallengeRepository.findByWeekStartLessThanEqualAndWeekEndGreaterThanEqual(
                any(Date.class), any(Date.class)))
                .thenReturn(Optional.of(challenge));

        UserChallengeProgress progress = new UserChallengeProgress();
        progress.setCompletedSessions(2);
        progress.setIsCompleted(false);
        when(userChallengeProgressRepository.findByUserIdAndChallengeId(1, 1))
                .thenReturn(Optional.of(progress));

        WeeklyChallengeDto result = getCurrentChallengeUseCase.execute(1);

        assertNotNull(result);
        assertEquals("コミュニケーション強化週間", result.title());
        assertEquals(2, result.completedSessions());
        assertFalse(result.isCompleted());
        assertEquals(3, result.targetSessions());
    }

    @Test
    @DisplayName("今週のチャレンジが存在しない場合、nullを返す")
    void execute_withoutChallenge_returnsNull() {
        when(weeklyChallengeRepository.findByWeekStartLessThanEqualAndWeekEndGreaterThanEqual(
                any(Date.class), any(Date.class)))
                .thenReturn(Optional.empty());

        WeeklyChallengeDto result = getCurrentChallengeUseCase.execute(1);

        assertNull(result);
    }
}
