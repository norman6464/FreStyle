package com.normanblog.frestyle.dto;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.normanblog.frestyle.entity.MasterExercise;
import java.time.Instant;

/**
 * GET /exercises のレスポンス 1 件。master 演習 + current user の status + 集計 stats を平坦に持つ
 * (frontend の MasterExerciseWithStatus extends MasterExercise に合わせフィールドを展開)。
 */
@JsonInclude(JsonInclude.Include.NON_NULL)
public record MasterExerciseWithStatusResponse(
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
    Instant updatedAt,
    String status,
    ExerciseStats stats) {

  /** 演習ごとの集計。 */
  public record ExerciseStats(long totalSubmissions, long solvedUsers) {}

  public static MasterExerciseWithStatusResponse of(
      MasterExercise e, String status, long totalSubmissions, long solvedUsers) {
    return new MasterExerciseWithStatusResponse(
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
        e.getUpdatedAt(),
        status,
        new ExerciseStats(totalSubmissions, solvedUsers));
  }
}
