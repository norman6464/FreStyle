package com.normanblog.frestyle.dto;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.normanblog.frestyle.entity.AiChatSession;
import java.time.Instant;

/** AI チャットセッション 1 件のクライアント向け表現。Go 版の JSON 形と互換のフィールド名にしている。 */
@JsonInclude(JsonInclude.Include.NON_NULL)
public record AiChatSessionResponse(
    Long id,
    Long userId,
    String title,
    String sessionType,
    // free のときは null(Go では omitempty)。
    Long scenarioId,
    Instant createdAt,
    Instant updatedAt) {

  public static AiChatSessionResponse from(AiChatSession s) {
    return new AiChatSessionResponse(
        s.getId(),
        s.getUserId(),
        s.getTitle(),
        s.getSessionType(),
        s.getScenarioId(),
        s.getCreatedAt(),
        s.getUpdatedAt());
  }
}
