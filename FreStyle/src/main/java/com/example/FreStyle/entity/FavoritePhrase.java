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

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Entity
@Table(name = "favorite_phrases")
@Data
@NoArgsConstructor
@AllArgsConstructor
public class FavoritePhrase {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Integer id;

    @ManyToOne
    @JoinColumn(name = "user_id", nullable = false)
    private User user;

    @Column(name = "original_text", nullable = false, columnDefinition = "TEXT")
    private String originalText;

    @Column(name = "rephrased_text", nullable = false, columnDefinition = "TEXT")
    private String rephrasedText;

    @Column(name = "pattern", nullable = false, length = 50)
    private String pattern;

    @Column(name = "created_at", insertable = false, updatable = false)
    private Timestamp createdAt;
}
