package com.normanblog.frestyle.dto;

import jakarta.validation.constraints.Size;
import java.util.List;

/**
 * AI チャット SSE ストリーミングのリクエスト。
 *
 * <p>sessionId が 0/null なら新規セッションを作る。content か attachments のいずれかは必須
 * (検証は usecase 側)。attachments は S3 へアップロード済みの参照(key + メタ)で、実体は backend が取得。
 */
public record AiChatStreamRequest(
    Long sessionId,
    String content,
    String sessionType,
    Long scenarioId,
    @Size(max = 4, message = "添付は最大 4 件です") List<AiChatStreamAttachment> attachments) {

  /** S3 アップロード済み添付の参照。 */
  public record AiChatStreamAttachment(
      String key, String filename, String contentType, long sizeBytes) {}
}
