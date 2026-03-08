package com.example.FreStyle.entity;

import java.sql.Date;
import java.sql.Timestamp;
import jakarta.persistence.*;
import lombok.*;

@Entity
@Table(name = "weekly_challenges")
@Data
@NoArgsConstructor
@AllArgsConstructor
public class WeeklyChallenge {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Integer id;

    @Column(length = 100, nullable = false)
    private String title;

    @Column(columnDefinition = "TEXT", nullable = false)
    private String description;

    @Column(length = 50, nullable = false)
    private String category;

    @Column(name = "target_sessions", nullable = false)
    private Integer targetSessions;

    @Column(name = "week_start", nullable = false)
    private Date weekStart;

    @Column(name = "week_end", nullable = false)
    private Date weekEnd;

    @Column(name = "created_at", insertable = false, updatable = false)
    private Timestamp createdAt;
}
