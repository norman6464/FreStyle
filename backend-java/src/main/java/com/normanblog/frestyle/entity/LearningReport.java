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

/** ユーザーの学習レポート(週次・月次集計を非同期で生成)を表す JPA エンティティ。 */
@Entity
@Table(name = "learning_reports")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class LearningReport {

  @Id
  @GeneratedValue(strategy = GenerationType.IDENTITY)
  private Long id;

  @Column(name = "user_id", nullable = false)
  private Long userId;

  @Column(name = "period_from", nullable = false)
  private Instant periodFrom;

  @Column(name = "period_to", nullable = false)
  private Instant periodTo;

  @Column(nullable = false)
  private String status;

  // 生成完了後の成果物の S3 キー(未完了の間は null)。
  @Column(name = "s3_key")
  private String s3Key;

  @Column(name = "created_at")
  private Instant createdAt;
}
