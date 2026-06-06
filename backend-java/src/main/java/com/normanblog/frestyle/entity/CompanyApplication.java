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
 * 未登録の企業担当者がログイン前に出す「利用申請」。
 *
 * <p>super_admin が一覧で確認し、問題なければ招待フローで company_admin を招待する入口になる。
 */
@Entity
@Table(name = "company_applications")
@Getter
@Setter
@NoArgsConstructor
@AllArgsConstructor
@Builder
public class CompanyApplication {

  @Id
  @GeneratedValue(strategy = GenerationType.IDENTITY)
  private Long id;

  @Column(name = "company_name", nullable = false)
  private String companyName;

  @Column(name = "applicant_name", nullable = false)
  private String applicantName;

  @Column(nullable = false)
  private String email;

  @Column(columnDefinition = "text")
  private String message;

  @Column(nullable = false)
  private String status;

  @Column(name = "created_at")
  private Instant createdAt;

  @Column(name = "updated_at")
  private Instant updatedAt;
}
