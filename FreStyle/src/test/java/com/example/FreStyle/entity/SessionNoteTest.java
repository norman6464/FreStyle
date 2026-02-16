package com.example.FreStyle.entity;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;

import java.time.LocalDateTime;

import static org.assertj.core.api.Assertions.assertThat;

class SessionNoteTest {

    @Test
    @DisplayName("updateTimestampでupdatedAtが現在時刻に設定される")
    void updateTimestampSetsUpdatedAt() {
        SessionNote note = new SessionNote();
        assertThat(note.getUpdatedAt()).isNull();

        LocalDateTime before = LocalDateTime.now();
        note.updateTimestamp();

        assertThat(note.getUpdatedAt()).isNotNull();
        assertThat(note.getUpdatedAt()).isAfterOrEqualTo(before);
        assertThat(note.getUpdatedAt()).isBeforeOrEqualTo(LocalDateTime.now());
    }

    @Test
    @DisplayName("updateTimestampを2回呼ぶとupdatedAtが更新される")
    void updateTimestampUpdatesOnSecondCall() throws InterruptedException {
        SessionNote note = new SessionNote();
        note.updateTimestamp();
        LocalDateTime firstTimestamp = note.getUpdatedAt();

        Thread.sleep(10);
        note.updateTimestamp();

        assertThat(note.getUpdatedAt()).isAfterOrEqualTo(firstTimestamp);
    }

    @Test
    @DisplayName("AllArgsConstructorで全フィールドが設定される")
    void allArgsConstructorWorks() {
        User user = new User();
        user.setId(1);
        LocalDateTime now = LocalDateTime.now();

        SessionNote note = new SessionNote(1, user, 100, "テストメモ", now);

        assertThat(note.getId()).isEqualTo(1);
        assertThat(note.getUser().getId()).isEqualTo(1);
        assertThat(note.getSessionId()).isEqualTo(100);
        assertThat(note.getNote()).isEqualTo("テストメモ");
        assertThat(note.getUpdatedAt()).isEqualTo(now);
    }
}
