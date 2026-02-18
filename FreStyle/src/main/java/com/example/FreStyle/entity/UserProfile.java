package com.example.FreStyle.entity;

import java.sql.Timestamp;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.FetchType;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.JoinColumn;
import jakarta.persistence.OneToOne;
import jakarta.persistence.Table;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import org.hibernate.annotations.JdbcTypeCode;
import org.hibernate.type.SqlTypes;

@Entity
@Table(name = "user_profiles")
@Data
@AllArgsConstructor
@NoArgsConstructor
public class UserProfile {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Integer id;

    @OneToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "user_id", nullable = false, unique = true)
    private User user;

    // 基本的な自己紹介
    @Column(name = "display_name", length = 100)
    private String displayName;

    @Column(name = "self_introduction", columnDefinition = "TEXT")
    private String selfIntroduction;

    // コミュニケーションスタイル
    @Column(name = "communication_style", length = 50)
    private String communicationStyle;

    @JdbcTypeCode(SqlTypes.JSON)
    @Column(name = "personality_traits", columnDefinition = "JSON")
    private String personalityTraits;

    // AIフィードバック用の追加情報
    @Column(columnDefinition = "TEXT")
    private String goals;

    @Column(columnDefinition = "TEXT")
    private String concerns;

    @Column(name = "preferred_feedback_style", length = 50)
    private String preferredFeedbackStyle;

    // メタ情報
    @Column(name = "created_at", insertable = false, updatable = false)
    private Timestamp createdAt;

    @Column(name = "updated_at", insertable = false, updatable = false)
    private Timestamp updatedAt;
}
