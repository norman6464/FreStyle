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
    @DisplayName("updateTimestampで過去のupdatedAtが現在時刻以降に更新される")
    void updateTimestampUpdatesFromPastTimestamp() {
        ScoreGoal scoreGoal = new ScoreGoal();
        LocalDateTime past = LocalDateTime.now().minusDays(1);
        scoreGoal.setUpdatedAt(past);

        scoreGoal.updateTimestamp();

        assertThat(scoreGoal.getUpdatedAt()).isAfter(past);
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
