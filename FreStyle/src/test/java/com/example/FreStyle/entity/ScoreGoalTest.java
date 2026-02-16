package com.example.FreStyle.entity;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.time.LocalDateTime;

import static org.assertj.core.api.Assertions.assertThat;

class ScoreGoalTest {

    @Test
    @DisplayName("updateTimestampでupdatedAtが現在時刻に設定される")
    void updateTimestampSetsUpdatedAt() {
        ScoreGoal scoreGoal = new ScoreGoal();
        assertThat(scoreGoal.getUpdatedAt()).isNull();

        LocalDateTime before = LocalDateTime.now();
        scoreGoal.updateTimestamp();

        assertThat(scoreGoal.getUpdatedAt()).isNotNull();
        assertThat(scoreGoal.getUpdatedAt()).isAfterOrEqualTo(before);
        assertThat(scoreGoal.getUpdatedAt()).isBeforeOrEqualTo(LocalDateTime.now());
    }

    @Test
    @DisplayName("updateTimestampを2回呼ぶとupdatedAtが更新される")
    void updateTimestampUpdatesOnSecondCall() throws InterruptedException {
        ScoreGoal scoreGoal = new ScoreGoal();
        scoreGoal.updateTimestamp();
        LocalDateTime firstTimestamp = scoreGoal.getUpdatedAt();

        Thread.sleep(10);
        scoreGoal.updateTimestamp();

        assertThat(scoreGoal.getUpdatedAt()).isAfterOrEqualTo(firstTimestamp);
    }

    @Test
    @DisplayName("AllArgsConstructorで全フィールドが設定される")
    void allArgsConstructorWorks() {
        User user = new User();
        user.setId(1);
        LocalDateTime now = LocalDateTime.now();

        ScoreGoal scoreGoal = new ScoreGoal(1, user, 8.5, now);

        assertThat(scoreGoal.getId()).isEqualTo(1);
        assertThat(scoreGoal.getUser().getId()).isEqualTo(1);
        assertThat(scoreGoal.getGoalScore()).isEqualTo(8.5);
        assertThat(scoreGoal.getUpdatedAt()).isEqualTo(now);
    }
}
