package com.normanblog.frestyle.dto;

import com.normanblog.frestyle.entity.Notification;
import java.time.Instant;

/** 通知 1 件のクライアント向け表現。Go 版の JSON 形と互換のフィールド名にしている。 */
public record NotificationResponse(
    Long id, Long userId, String type, String title, String body, boolean isRead, Instant createdAt) {

  public static NotificationResponse from(Notification n) {
    return new NotificationResponse(
        n.getId(),
        n.getUserId(),
        n.getType(),
        n.getTitle(),
        n.getBody(),
        n.isRead(),
        n.getCreatedAt());
  }
}
