package com.normanblog.frestyle.service;

import java.util.Map;

/**
 * AI チャット添付の許容 MIME とそのメタ(Bedrock format / kind / サイズ上限)。
 *
 * <p>署名 URL 発行(検証)と SSE 送信(添付実体の組み立て)の両方で参照する単一情報源。現状は画像のみ
 * 許可(Go 版と同値)。
 */
public final class AiChatAttachmentRules {

  /** 許容 MIME に対応するルール。format は Bedrock の "png"/"jpeg" 等、kind は "image"/"document"。 */
  public record Rule(String format, String kind, long maxBytes) {}

  private static final long MAX_IMAGE_BYTES = 5L * 1024 * 1024;

  private static final Map<String, Rule> RULES =
      Map.of(
          "image/png", new Rule("png", "image", MAX_IMAGE_BYTES),
          "image/jpeg", new Rule("jpeg", "image", MAX_IMAGE_BYTES),
          "image/jpg", new Rule("jpeg", "image", MAX_IMAGE_BYTES),
          "image/gif", new Rule("gif", "image", MAX_IMAGE_BYTES),
          "image/webp", new Rule("webp", "image", MAX_IMAGE_BYTES));

  private AiChatAttachmentRules() {}

  /** contentType のルールを返す(未対応なら null)。 */
  public static Rule of(String contentType) {
    return contentType == null ? null : RULES.get(contentType);
  }

  /** 許容 MIME 一覧。 */
  public static java.util.Set<String> allowedContentTypes() {
    return RULES.keySet();
  }
}
