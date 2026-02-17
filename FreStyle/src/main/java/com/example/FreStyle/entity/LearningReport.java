package com.example.FreStyle.entity;

import java.time.LocalDateTime;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.JoinColumn;
import jakarta.persistence.ManyToOne;
import jakarta.persistence.PrePersist;
import jakarta.persistence.Table;
import jakarta.persistence.UniqueConstraint;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Entity
@Table(name = "learning_reports", uniqueConstraints = @UniqueConstraint(columnNames = {"user_id", "year", "month"}))
@Data
@NoArgsConstructor
@AllArgsConstructor
public class LearningReport {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Integer id;

    @ManyToOne
    @JoinColumn(name = "user_id", nullable = false)
    private User user;

    @Column(name = "year", nullable = false)
    private Integer year;

    @Column(name = "month", nullable = false)
    private Integer month;

    @Column(name = "total_sessions", nullable = false)
    private Integer totalSessions;

    @Column(name = "average_score", nullable = false)
    private Double averageScore;

    @Column(name = "previous_average_score")
    private Double previousAverageScore;

    @Column(name = "best_axis", length = 50)
    private String bestAxis;

    @Column(name = "worst_axis", length = 50)
    private String worstAxis;

    @Column(name = "practice_days", nullable = false)
    private Integer practiceDays;

    @Column(name = "created_at", nullable = false)
    private LocalDateTime createdAt;

    @PrePersist
    public void prePersist() {
        if (this.createdAt == null) {
            this.createdAt = LocalDateTime.now();
        }
    }
}
