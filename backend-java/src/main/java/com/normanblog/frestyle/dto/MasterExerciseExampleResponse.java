package com.normanblog.frestyle.dto;

import com.normanblog.frestyle.entity.MasterExerciseExample;
import java.time.Instant;

/** master 演習の入力/期待出力の例 1 件。 */
public record MasterExerciseExampleResponse(
    Long id,
    Long exerciseId,
    short orderIndex,
    String inputText,
    String expectedOutput,
    Instant createdAt,
    Instant updatedAt) {

  public static MasterExerciseExampleResponse from(MasterExerciseExample e) {
    return new MasterExerciseExampleResponse(
        e.getId(),
        e.getExerciseId(),
        e.getOrderIndex(),
        e.getInputText(),
        e.getExpectedOutput(),
        e.getCreatedAt(),
        e.getUpdatedAt());
  }
}
