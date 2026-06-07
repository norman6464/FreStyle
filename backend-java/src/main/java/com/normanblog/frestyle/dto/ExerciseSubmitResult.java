package com.normanblog.frestyle.dto;

import java.util.List;

/** POST /exercises/{slug}/submit のレスポンス。採点結果(テストケース別)+ 正誤。 */
public record ExerciseSubmitResult(
    Long submissionId, boolean isCorrect, List<TestCaseResult> results) {

  /** テストケース 1 件の採点結果。 */
  public record TestCaseResult(
      short orderIndex,
      String input,
      String expectedOutput,
      String actualOutput,
      String stderr,
      boolean passed) {}
}
