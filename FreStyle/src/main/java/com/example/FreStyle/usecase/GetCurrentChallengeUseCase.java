package com.example.FreStyle.usecase;

import java.sql.Date;
import java.time.LocalDate;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.WeeklyChallengeDto;
import com.example.FreStyle.entity.WeeklyChallenge;
import com.example.FreStyle.repository.UserChallengeProgressRepository;
import com.example.FreStyle.repository.WeeklyChallengeRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class GetCurrentChallengeUseCase {

    private final WeeklyChallengeRepository weeklyChallengeRepository;
    private final UserChallengeProgressRepository userChallengeProgressRepository;

    @Transactional(readOnly = true)
    public WeeklyChallengeDto execute(Integer userId) {
        Date today = Date.valueOf(LocalDate.now());
        return weeklyChallengeRepository
                .findByWeekStartLessThanEqualAndWeekEndGreaterThanEqual(today, today)
                .map(challenge -> {
                    var progress = userChallengeProgressRepository
                            .findByUserIdAndChallengeId(userId, challenge.getId());
                    int completedSessions = progress.map(p -> p.getCompletedSessions()).orElse(0);
                    boolean isCompleted = progress.map(p -> p.getIsCompleted()).orElse(false);
                    return toDto(challenge, completedSessions, isCompleted);
                })
                .orElse(null);
    }

    private WeeklyChallengeDto toDto(WeeklyChallenge challenge, int completedSessions, boolean isCompleted) {
        return new WeeklyChallengeDto(
                challenge.getId(),
                challenge.getTitle(),
                challenge.getDescription(),
                challenge.getCategory(),
                challenge.getTargetSessions(),
                completedSessions,
                isCompleted,
                challenge.getWeekStart().toString(),
                challenge.getWeekEnd().toString());
    }
}
