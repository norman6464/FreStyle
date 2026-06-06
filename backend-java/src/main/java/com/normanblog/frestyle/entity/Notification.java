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

/** ユーザー向けの通知を表す JPA エンティティ。 */
@Entity
@Table(name = "notifications")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class Notification {

  @Id
  @GeneratedValue(strategy = GenerationType.IDENTITY)
  private Long id;

  @Column(name = "user_id", nullable = false)
  private Long userId;

  @Column(nullable = false)
  private String type;

  @Column(nullable = false)
  private String title;

  @Column(columnDefinition = "text", nullable = false)
  private String body;

  @Column(name = "is_read", nullable = false)
  private boolean isRead;

  @Column(name = "created_at")
  private Instant createdAt;
}
