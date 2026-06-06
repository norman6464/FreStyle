package com.normanblog.frestyle.entity;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import java.time.Instant;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

/**
 * users とは別管理のプロフィール拡張情報。主キーは users.id と一致する user_id(1:1)。
 *
 * <p>displayName / email は users 側に持つため、ここには bio / avatar / status だけを持つ。
 */
@Entity
@Table(name = "profiles")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class Profile {

  // users.id と同値。自動採番せず、ユーザーの id を明示的に入れる。
  @Id
  @Column(name = "user_id")
  private Long userId;

  @Column(columnDefinition = "text")
  private String bio;

  @Column(name = "avatar_url", columnDefinition = "text")
  private String avatarUrl;

  @Column(columnDefinition = "text")
  private String status;

  @Column(name = "updated_at")
  private Instant updatedAt;
}
