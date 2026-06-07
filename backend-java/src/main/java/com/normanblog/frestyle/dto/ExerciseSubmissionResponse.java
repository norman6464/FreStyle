package com.normanblog.frestyle.dto;

import com.normanblog.frestyle.entity.ExerciseSubmission;
import java.time.Instant;

/** GET /exercises/{slug}/submissions の 1 件。 */
public record ExerciseSubmissionResponse(
    Long id,
    Long userId,
    String exerciseKind,
    Long exerciseId,
    String submittedCode,
    String stdout,
    String stderr,
    int exitCode,
    boolean isCorrect,
    Instant submittedAt) {

  public static ExerciseSubmissionResponse from(ExerciseSubmission s) {
    return new ExerciseSubmissionResponse(
        s.getId(),
        s.getUserId(),
        s.getExerciseKind(),
        s.getExerciseId(),
        s.getSubmittedCode(),
        s.getStdout(),
        s.getStderr(),
        s.getExitCode(),
        s.isCorrect(),
        s.getSubmittedAt());
  }
}
