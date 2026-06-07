package com.normanblog.frestyle.entity;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import java.time.Instant;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

/** AI チャットの 1 セッション(メタデータ)を表す JPA エンティティ。メッセージ本体は DynamoDB。 */
@Entity
@Table(name = "ai_chat_sessions")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class AiChatSession {

  @Id
  @GeneratedValue(strategy = GenerationType.IDENTITY)
  private Long id;

  @Column(name = "user_id", nullable = false)
  private Long userId;

  @Column(nullable = false)
  private String title;

  @Column(name = "session_type", nullable = false)
  private String sessionType;

  // practice 用シナリオ ID(free のときは null)。
  @Column(name = "scenario_id")
  private Long scenarioId;

  @Column(name = "created_at")
  private Instant createdAt;

  @Column(name = "updated_at")
  private Instant updatedAt;
}
