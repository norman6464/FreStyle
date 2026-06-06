package com.normanblog.frestyle.service;

import com.normanblog.frestyle.entity.Notification;
import com.normanblog.frestyle.repository.NotificationRepository;
import java.time.Instant;
import java.util.List;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

/** 通知の一覧取得・既読化・未読数集計を担うサービス。 */
@Service
public class NotificationService {

  private final NotificationRepository notifications;

  public NotificationService(NotificationRepository notifications) {
    this.notifications = notifications;
  }

  /** 通知を 1 件作成する(システムがユーザーへ通知を出す用途)。 */
  public Notification create(Long userId, String type, String title, String body) {
    Notification notification =
        Notification.builder()
            .userId(userId)
            .type(type)
            .title(title)
            .body(body)
            .isRead(false)
            .createdAt(Instant.now())
            .build();

    return notifications.save(notification);
  }

  /** current user の通知を作成日降順で返す。 */
  public List<Notification> list(Long userId) {
    return notifications.findByUserIdOrderByCreatedAtDesc(userId);
  }

  /** 指定通知を既読化する。所有者検証は WHERE 条件で担保するため他人の通知は更新されない。 */
  @Transactional
  public void markRead(Long userId, Long id) {
    notifications.markRead(userId, id);
  }

  /** current user の全未読通知をまとめて既読化する。 */
  @Transactional
  public void markAllRead(Long userId) {
    notifications.markAllRead(userId);
  }

  /** current user の未読通知数を返す(バッジ表示用)。 */
  public long countUnread(Long userId) {
    return notifications.countByUserIdAndIsReadFalse(userId);
  }
}
