package com.example.FreStyle.entity;

import java.sql.Timestamp;
import jakarta.persistence.*;
import lombok.*;

@Entity
@Table(name = "user_challenge_progress")
@Data
@NoArgsConstructor
@AllArgsConstructor
public class UserChallengeProgress {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Integer id;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "user_id", nullable = false)
    private User user;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "challenge_id", nullable = false)
    private WeeklyChallenge challenge;

    @Column(name = "completed_sessions", nullable = false)
    private Integer completedSessions;

    @Column(name = "is_completed", nullable = false)
    private Boolean isCompleted;

    @Column(name = "updated_at", insertable = false, updatable = false)
    private Timestamp updatedAt;
}
