package com.example.FreStyle.usecase;

import java.sql.Date;
import java.time.LocalDate;

import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import com.example.FreStyle.dto.WeeklyChallengeDto;
import com.example.FreStyle.entity.UserChallengeProgress;
import com.example.FreStyle.entity.WeeklyChallenge;
import com.example.FreStyle.repository.UserChallengeProgressRepository;
import com.example.FreStyle.repository.WeeklyChallengeRepository;

import lombok.RequiredArgsConstructor;

@Service
@RequiredArgsConstructor
public class IncrementChallengeProgressUseCase {

    private final WeeklyChallengeRepository weeklyChallengeRepository;
    private final UserChallengeProgressRepository userChallengeProgressRepository;

    @Transactional
    public WeeklyChallengeDto execute(Integer userId) {
        Date today = Date.valueOf(LocalDate.now());
        WeeklyChallenge challenge = weeklyChallengeRepository
                .findByWeekStartLessThanEqualAndWeekEndGreaterThanEqual(today, today)
                .orElseThrow(() -> new RuntimeException("今週のチャレンジが見つかりません"));

        UserChallengeProgress progress = userChallengeProgressRepository
                .findByUserIdAndChallengeId(userId, challenge.getId())
                .orElseGet(() -> {
                    UserChallengeProgress newProgress = new UserChallengeProgress();
                    newProgress.setChallenge(challenge);
                    newProgress.setCompletedSessions(0);
                    newProgress.setIsCompleted(false);
                    return newProgress;
                });

        progress.setCompletedSessions(progress.getCompletedSessions() + 1);
        if (progress.getCompletedSessions() >= challenge.getTargetSessions()) {
            progress.setIsCompleted(true);
        }
        userChallengeProgressRepository.save(progress);

        return new WeeklyChallengeDto(
                challenge.getId(),
                challenge.getTitle(),
                challenge.getDescription(),
                challenge.getCategory(),
                challenge.getTargetSessions(),
                progress.getCompletedSessions(),
                progress.getIsCompleted(),
                challenge.getWeekStart().toString(),
                challenge.getWeekEnd().toString());
    }
}
