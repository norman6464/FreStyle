package com.normanblog.frestyle.dto;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.normanblog.frestyle.entity.LearningReport;
import java.time.Instant;

/** 学習レポート 1 件のクライアント向け表現。Go 版の JSON 形と互換のフィールド名にしている。 */
@JsonInclude(JsonInclude.Include.NON_NULL)
public record LearningReportResponse(
    Long id,
    Long userId,
    Instant periodFrom,
    Instant periodTo,
    String status,
    // 未完了の間は null。Go では omitempty のため null は省略する。
    String s3Key,
    Instant createdAt) {

  public static LearningReportResponse from(LearningReport r) {
    return new LearningReportResponse(
        r.getId(),
        r.getUserId(),
        r.getPeriodFrom(),
        r.getPeriodTo(),
        r.getStatus(),
        r.getS3Key(),
        r.getCreatedAt());
  }
}
