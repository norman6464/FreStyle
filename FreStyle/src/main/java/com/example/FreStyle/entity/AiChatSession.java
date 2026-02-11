package com.example.FreStyle.entity;

import java.sql.Timestamp;
import java.util.List;

import jakarta.persistence.CascadeType;
import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.FetchType;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.JoinColumn;
import jakarta.persistence.ManyToOne;
import jakarta.persistence.OneToMany;
import jakarta.persistence.Table;
import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Entity
@Table(name = "ai_chat_sessions")
@Data
@NoArgsConstructor
@AllArgsConstructor
public class AiChatSession {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Integer id;

    // セッションの所有者
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "user_id", nullable = false)
    private User user;

    // セッションのタイトル（自動生成 or ユーザー設定）
    @Column(name = "title", length = 255)
    private String title;

    // chat_roomsとの関連（会話レビューの場合）
    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "related_room_id")
    private ChatRoom relatedRoom;

    // フィードバック時のシーン（meeting, one_on_one, email, presentation, negotiation）
    @Column(name = "scene", length = 50)
    private String scene;

    @Column(name = "created_at", insertable = false, updatable = false)
    private Timestamp createdAt;

    @Column(name = "updated_at", insertable = false, updatable = false)
    private Timestamp updatedAt;

    // このセッションに属するメッセージ一覧
    @OneToMany(mappedBy = "session", cascade = CascadeType.ALL, orphanRemoval = true)
    private List<AiChatMessage> messages;

}
