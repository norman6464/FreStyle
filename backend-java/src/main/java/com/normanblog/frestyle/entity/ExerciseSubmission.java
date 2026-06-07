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

/**
 * trainee がコード演習に提出したコード + 実行結果の 1 件(append-only)。
 *
 * <p>参照先は exerciseKind で判定する polymorphic 設計(FK は張らずアプリ層で担保)。
 */
@Entity
@Table(name = "exercise_submissions")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class ExerciseSubmission {

  @Id
  @GeneratedValue(strategy = GenerationType.IDENTITY)
  private Long id;

  @Column(name = "user_id", nullable = false)
  private Long userId;

  @Column(name = "exercise_kind", nullable = false)
  private String exerciseKind;

  @Column(name = "exercise_id", nullable = false)
  private Long exerciseId;

  @Column(name = "submitted_code", columnDefinition = "text", nullable = false)
  private String submittedCode;

  @Column(columnDefinition = "text")
  private String stdout;

  @Column(columnDefinition = "text")
  private String stderr;

  @Column(name = "exit_code", nullable = false)
  private int exitCode;

  @Column(name = "is_correct", nullable = false)
  private boolean isCorrect;

  @Column(name = "submitted_at", nullable = false)
  private Instant submittedAt;
}
