package com.normanblog.frestyle.entity;

/** ai_chat_sessions.session_type の許容値。DB には文字列で保存する(既存 Go 実装と同値)。 */
public final class AiChatSessionType {

  // 自由会話。
  public static final String FREE = "free";
  // シナリオ練習。
  public static final String PRACTICE = "practice";

  private AiChatSessionType() {}
}
