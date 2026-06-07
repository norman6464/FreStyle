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

/** 運営が用意した練習問題のマスタ。language で言語を表現し言語非依存に扱う。 */
@Entity
@Table(name = "master_exercises")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class MasterExercise {

  @Id
  @GeneratedValue(strategy = GenerationType.IDENTITY)
  private Long id;

  @Column(nullable = false, unique = true)
  private String slug;

  @Column(nullable = false)
  private String language;

  @Column(name = "order_index", nullable = false)
  private int orderIndex;

  @Column(nullable = false)
  private String category;

  @Column(nullable = false)
  private String title;

  @Column(columnDefinition = "text", nullable = false)
  private String description;

  @Column(name = "starter_code", columnDefinition = "text", nullable = false)
  private String starterCode;

  @Column(name = "hint_text", columnDefinition = "text")
  private String hintText;

  @Column(name = "expected_output", columnDefinition = "text")
  private String expectedOutput;

  // 採点モード: execute(実行して stdout 比較) / qa(提出文字列と expectedOutput を比較)。
  @Column(nullable = false)
  private String mode;

  @Column(columnDefinition = "text", nullable = false)
  private String explanation;

  @Column(nullable = false)
  private short difficulty;

  @Column(name = "is_published", nullable = false)
  private boolean isPublished;

  @Column(name = "chapter_id")
  private Long chapterId;

  @Column(name = "created_at")
  private Instant createdAt;

  @Column(name = "updated_at")
  private Instant updatedAt;
}
