package com.normanblog.frestyle.repository;

import com.normanblog.frestyle.entity.Notification;
import java.util.List;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

/** notifications テーブルへのアクセスを担うリポジトリ。 */
public interface NotificationRepository extends JpaRepository<Notification, Long> {

  List<Notification> findByUserIdOrderByCreatedAtDesc(Long userId);

  long countByUserIdAndIsReadFalse(Long userId);

  // WHERE で user_id を絞り、他人の通知を既読化できないようにする(所有者検証)。
  @Modifying
  @Query("update Notification n set n.isRead = true where n.id = :id and n.userId = :userId")
  int markRead(@Param("userId") Long userId, @Param("id") Long id);

  @Modifying
  @Query("update Notification n set n.isRead = true where n.userId = :userId and n.isRead = false")
  int markAllRead(@Param("userId") Long userId);
}
