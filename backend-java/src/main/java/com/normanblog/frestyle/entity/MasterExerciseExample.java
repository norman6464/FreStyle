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

/** MasterExercise に紐付く入力例 / 期待出力例の 1 ペア。表示 / 採点順序は orderIndex で安定ソート。 */
@Entity
@Table(name = "master_exercise_examples")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class MasterExerciseExample {

  @Id
  @GeneratedValue(strategy = GenerationType.IDENTITY)
  private Long id;

  @Column(name = "exercise_id", nullable = false)
  private Long exerciseId;

  @Column(name = "order_index", nullable = false)
  private short orderIndex;

  @Column(name = "input_text", columnDefinition = "text", nullable = false)
  private String inputText;

  @Column(name = "expected_output", columnDefinition = "text", nullable = false)
  private String expectedOutput;

  @Column(name = "created_at")
  private Instant createdAt;

  @Column(name = "updated_at")
  private Instant updatedAt;
}
