package com.normanblog.frestyle.entity;

/** master_exercises.mode の許容値。 */
public final class ExerciseMode {

  // 実行して stdout を期待値と比較する。
  public static final String EXECUTE = "execute";
  // 提出文字列と expectedOutput を比較する(実行しない)。
  public static final String QA = "qa";

  private ExerciseMode() {}
}
