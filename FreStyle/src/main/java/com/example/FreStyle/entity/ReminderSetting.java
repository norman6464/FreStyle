package com.example.FreStyle.entity;

import java.sql.Timestamp;
import jakarta.persistence.*;
import lombok.*;

@Entity
@Table(name = "reminder_settings")
@Data
@NoArgsConstructor
@AllArgsConstructor
public class ReminderSetting {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Integer id;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "user_id", nullable = false)
    private User user;

    @Column(nullable = false)
    private Boolean enabled;

    @Column(name = "reminder_time", length = 5, nullable = false)
    private String reminderTime;

    @Column(name = "days_of_week", length = 50, nullable = false)
    private String daysOfWeek;

    @Column(name = "updated_at", insertable = false, updatable = false)
    private Timestamp updatedAt;
}
