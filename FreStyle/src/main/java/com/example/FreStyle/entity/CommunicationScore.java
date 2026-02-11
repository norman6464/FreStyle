package com.example.FreStyle.entity;

import java.sql.Timestamp;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.FetchType;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.JoinColumn;
import jakarta.persistence.ManyToOne;
import jakarta.persistence.Table;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Entity
@Table(name = "communication_scores")
@Data
@NoArgsConstructor
@AllArgsConstructor
public class CommunicationScore {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Integer id;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "session_id", nullable = false)
    private AiChatSession session;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "user_id", nullable = false)
    private User user;

    @Column(name = "axis_name", length = 50, nullable = false)
    private String axisName;

    @Column(name = "score", nullable = false)
    private Integer score;

    @Column(name = "comment", columnDefinition = "TEXT")
    private String comment;

    @Column(name = "scene", length = 50)
    private String scene;

    @Column(name = "created_at", insertable = false, updatable = false)
    private Timestamp createdAt;
}
