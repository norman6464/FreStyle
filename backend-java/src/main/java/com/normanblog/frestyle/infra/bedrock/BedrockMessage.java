package com.normanblog.frestyle.infra.bedrock;

import java.util.List;

/**
 * Bedrock Converse へ渡す会話 1 件分の中立モデル(AWS SDK 型を上位層に漏らさない)。
 *
 * <p>role は "user" / "assistant"。images は最新ユーザー発話にのみ実体(bytes)が入り、過去履歴は空。
 */
public record BedrockMessage(String role, String text, List<BedrockImage> images) {

  /** 添付画像 1 枚。format は Bedrock の "png" / "jpeg" 等。 */
  public record BedrockImage(String format, byte[] bytes) {}
}
