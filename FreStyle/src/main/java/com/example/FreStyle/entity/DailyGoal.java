package com.example.FreStyle.entity;

import java.time.LocalDate;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.JoinColumn;
import jakarta.persistence.ManyToOne;
import jakarta.persistence.Table;
import jakarta.persistence.UniqueConstraint;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Entity
@Table(name = "daily_goals", uniqueConstraints = @UniqueConstraint(columnNames = {"user_id", "goal_date"}))
@Data
@NoArgsConstructor
@AllArgsConstructor
public class DailyGoal {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Integer id;

    @ManyToOne
    @JoinColumn(name = "user_id", nullable = false)
    private User user;

    @Column(name = "goal_date", nullable = false)
    private LocalDate goalDate;

    @Column(name = "target", nullable = false)
    private Integer target;

    @Column(name = "completed", nullable = false)
    private Integer completed;
}
