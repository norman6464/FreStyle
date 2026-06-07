package com.normanblog.frestyle.dto;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.normanblog.frestyle.entity.MasterExercise;
import java.time.Instant;

/** master 演習 1 件のクライアント向け表現(Go 版 domain.MasterExercise と互換のフィールド名)。 */
@JsonInclude(JsonInclude.Include.NON_NULL)
public record MasterExerciseResponse(
    Long id,
    String slug,
    String language,
    int orderIndex,
    String category,
    String title,
    String description,
    String starterCode,
    String hintText,
    String expectedOutput,
    String mode,
    String explanation,
    short difficulty,
    boolean isPublished,
    Long chapterId,
    Instant createdAt,
    Instant updatedAt) {

  public static MasterExerciseResponse from(MasterExercise e) {
    return new MasterExerciseResponse(
        e.getId(),
        e.getSlug(),
        e.getLanguage(),
        e.getOrderIndex(),
        e.getCategory(),
        e.getTitle(),
        e.getDescription(),
        e.getStarterCode(),
        e.getHintText(),
        e.getExpectedOutput(),
        e.getMode(),
        e.getExplanation(),
        e.getDifficulty(),
        e.isPublished(),
        e.getChapterId(),
        e.getCreatedAt(),
        e.getUpdatedAt());
  }
}
