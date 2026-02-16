package com.example.FreStyle.entity;

import java.sql.Timestamp;

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
@Table(name = "scenario_bookmarks", uniqueConstraints = {
    @UniqueConstraint(columnNames = {"user_id", "scenario_id"})
})
@Data
@NoArgsConstructor
@AllArgsConstructor
public class ScenarioBookmark {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Integer id;

    @ManyToOne
    @JoinColumn(name = "user_id", nullable = false)
    private User user;

    @ManyToOne
    @JoinColumn(name = "scenario_id", nullable = false)
    private PracticeScenario scenario;

    @Column(name = "created_at", insertable = false, updatable = false)
    private Timestamp createdAt;
}
