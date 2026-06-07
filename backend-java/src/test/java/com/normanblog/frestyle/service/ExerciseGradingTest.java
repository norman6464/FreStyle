package com.normanblog.frestyle.service;

import static org.assertj.core.api.Assertions.assertThat;

import org.junit.jupiter.api.Test;

/** 採点の出力正規化・一致判定を検証する(行末改行/CRLF/末尾空白の扱い)。 */
class ExerciseGradingTest {

  @Test
  void normalizeOutput_null_returnsEmpty() {
    assertThat(ExerciseGrading.normalizeOutput(null)).isEmpty();
  }

  @Test
  void normalizeOutput_stripsTrailingWhitespaceAndNewlines() {
    assertThat(ExerciseGrading.normalizeOutput("hello\n\n  ")).isEqualTo("hello");
    assertThat(ExerciseGrading.normalizeOutput("hello\t \n")).isEqualTo("hello");
  }

  @Test
  void normalizeOutput_unifiesCrlfAndCr() {
    assertThat(ExerciseGrading.normalizeOutput("a\r\nb\rc")).isEqualTo("a\nb\nc");
  }

  @Test
  void normalizeOutput_keepsInnerWhitespace() {
    // 行内の空白は厳密一致(中間の空白は潰さない)。
    assertThat(ExerciseGrading.normalizeOutput("a  b\n")).isEqualTo("a  b");
  }

  @Test
  void matches_ignoresTrailingNewlineDifference() {
    assertThat(ExerciseGrading.matches("42\n", "42")).isTrue();
    assertThat(ExerciseGrading.matches("42\r\n", "42\n")).isTrue();
  }

  @Test
  void matches_detectsRealDifference() {
    assertThat(ExerciseGrading.matches("42", "43")).isFalse();
    assertThat(ExerciseGrading.matches("4 2", "42")).isFalse();
  }

  @Test
  void matches_bothNullOrEmpty() {
    assertThat(ExerciseGrading.matches(null, "")).isTrue();
    assertThat(ExerciseGrading.matches("", null)).isTrue();
  }
}
