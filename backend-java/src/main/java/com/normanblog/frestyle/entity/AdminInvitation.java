package com.normanblog.frestyle.entity;

import jakarta.persistence.Column;
import jakarta.persistence.Entity;
import jakarta.persistence.GeneratedValue;
import jakarta.persistence.GenerationType;
import jakarta.persistence.Id;
import jakarta.persistence.Table;
import java.time.Instant;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

/**
 * 企業担当者を招待するためのマジックリンク招待。
 *
 * <p>初回ログイン時にこの行(pending)があれば、新規ユーザーの作成を許可し、role / company を反映する。
 */
@Entity
@Table(name = "invitations")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class AdminInvitation {

  @Id
  @GeneratedValue(strategy = GenerationType.IDENTITY)
  private Long id;

  @Column(name = "company_id")
  private Long companyId;

  private String email;

  private String role;

  @Column(name = "display_name")
  private String displayName;

  private String status;

  // マジックリンク用の不透明トークン(秘匿値)。未設定を NULL にして UNIQUE 制約に当てないため null 許容。
  private String token;

  @Column(name = "expires_at")
  private Instant expiresAt;

  @Column(name = "created_at")
  private Instant createdAt;
}
