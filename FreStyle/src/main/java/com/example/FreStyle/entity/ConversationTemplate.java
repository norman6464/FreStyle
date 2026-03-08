package com.example.FreStyle.entity;

import java.sql.Timestamp;
import jakarta.persistence.*;
import lombok.*;

@Entity
@Table(name = "conversation_templates")
@Data
@NoArgsConstructor
@AllArgsConstructor
public class ConversationTemplate {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Integer id;

    @Column(name = "title", length = 100, nullable = false)
    private String title;

    @Column(name = "description", columnDefinition = "TEXT")
    private String description;

    @Column(name = "category", length = 50, nullable = false)
    private String category;

    @Column(name = "opening_message", columnDefinition = "TEXT", nullable = false)
    private String openingMessage;

    @Column(name = "system_context", columnDefinition = "TEXT", nullable = false)
    private String systemContext;

    @Column(name = "difficulty", length = 20)
    private String difficulty;

    @Column(name = "created_at", insertable = false, updatable = false)
    private Timestamp createdAt;
}
