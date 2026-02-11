package com.example.FreStyle.entity;

import java.sql.Timestamp;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Entity
@Table(name = "practice_scenarios")
@Data
@NoArgsConstructor
@AllArgsConstructor
public class PracticeScenario {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Integer id;

    @Column(name = "name", length = 100, nullable = false)
    private String name;

    @Column(name = "description", columnDefinition = "TEXT")
    private String description;

    @Column(name = "category", length = 50)
    private String category;

    @Column(name = "role_name", length = 100, nullable = false)
    private String roleName;

    @Column(name = "difficulty", length = 20)
    private String difficulty;

    @Column(name = "system_prompt", columnDefinition = "TEXT", nullable = false)
    private String systemPrompt;

    @Column(name = "created_at", insertable = false, updatable = false)
    private Timestamp createdAt;
}
