package com.normanblog.frestyle.service;

/**
 * 演習採点の出力正規化。CRLF/CR を LF に統一し、末尾の改行・空白を除去する(行内の空白は厳密一致)。
 *
 * <p>従来 frontend にも同等の関数(normalizeOutput)があったが、採点の正本はサーバ側に置く方が
 * テストできるため backend に一本化する。
 */
public final class ExerciseGrading {

  private ExerciseGrading() {}

  /** 出力を正規化する。 */
  public static String normalizeOutput(String s) {
    if (s == null) {
      return "";
    }
    String normalized = s.replace("\r\n", "\n").replace("\r", "\n");
    // 末尾の空白・タブ・改行を除去(Go の TrimRight(" \t\n") と同等)。
    int end = normalized.length();
    while (end > 0) {
      char c = normalized.charAt(end - 1);
      if (c == ' ' || c == '\t' || c == '\n') {
        end--;
      } else {
        break;
      }
    }
    return normalized.substring(0, end);
  }

  /** 2 つの出力が正規化後に一致するか。 */
  public static boolean matches(String actual, String expected) {
    return normalizeOutput(actual).equals(normalizeOutput(expected));
  }
}
