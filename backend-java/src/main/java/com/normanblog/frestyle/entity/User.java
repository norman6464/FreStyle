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
 * アプリ利用者を表す JPA エンティティ。
 *
 * <p>外部 ID(Cognito の {@code sub})と内部 ID({@code id})を分離した 2 層構成にしている。
 * 業務データ(notes 等)の所有者は内部 {@code id} で参照し、{@code cognitoSub} は認証時に
 * users 行を引くための橋渡しキーとしてのみ使う。これにより認証基盤(Cognito)を差し替えても
 * 業務データの外部キーを書き換えずに済む。
 */
@Entity
@Table(name = "users")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class User {

  @Id
  @GeneratedValue(strategy = GenerationType.IDENTITY)
  private Long id;

  @Column(name = "cognito_sub", unique = true)
  private String cognitoSub;

  private String email;

  @Column(name = "display_name")
  private String displayName;

  @Column(name = "company_id")
  private Long companyId;

  private String role;

  // Welcome 完了日時。null なら未オンボーディング(フロントの /welcome 判定に使う)。
  @Column(name = "onboarded_at")
  private Instant onboardedAt;

  @Column(name = "created_at")
  private Instant createdAt;

  @Column(name = "updated_at")
  private Instant updatedAt;

  // 論理削除。null でなければ削除済み。
  @Column(name = "deleted_at")
  private Instant deletedAt;
}
