package com.normanblog.frestyle.controller;

import com.normanblog.frestyle.dto.NotificationResponse;
import com.normanblog.frestyle.security.CurrentUserProvider;
import com.normanblog.frestyle.service.NotificationService;
import java.util.List;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PatchMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/** 通知 API のエンドポイントを公開するコントローラ。 */
@RestController
@RequestMapping("/api/v2/notifications")
public class NotificationController {

  private final NotificationService notifications;
  private final CurrentUserProvider currentUser;

  public NotificationController(
      NotificationService notifications, CurrentUserProvider currentUser) {
    this.notifications = notifications;
    this.currentUser = currentUser;
  }

  @GetMapping
  public List<NotificationResponse> list() {
    Long userId = currentUser.require().getId();
    return notifications.list(userId).stream().map(NotificationResponse::from).toList();
  }

  @GetMapping("/unread-count")
  public long unreadCount() {
    Long userId = currentUser.require().getId();
    return notifications.countUnread(userId);
  }

  // 旧クライアント互換で PUT も同じパスで受け付ける(標準は PATCH)。
  @PatchMapping("/{id}/read")
  @PutMapping("/{id}/read")
  public ResponseEntity<Void> markRead(@PathVariable Long id) {
    Long userId = currentUser.require().getId();
    notifications.markRead(userId, id);
    return ResponseEntity.noContent().build();
  }

  @PatchMapping("/read-all")
  @PutMapping("/read-all")
  public ResponseEntity<Void> markAllRead() {
    Long userId = currentUser.require().getId();
    notifications.markAllRead(userId);
    return ResponseEntity.noContent().build();
  }
}
