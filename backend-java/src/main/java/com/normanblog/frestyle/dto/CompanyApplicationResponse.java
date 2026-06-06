package com.normanblog.frestyle.dto;

import com.normanblog.frestyle.entity.CompanyApplication;
import java.time.Instant;

/** 企業利用申請の API レスポンス。 */
public record CompanyApplicationResponse(
    Long id,
    String companyName,
    String applicantName,
    String email,
    String message,
    String status,
    Instant createdAt,
    Instant updatedAt) {

  public static CompanyApplicationResponse from(CompanyApplication application) {
    return new CompanyApplicationResponse(
        application.getId(),
        application.getCompanyName(),
        application.getApplicantName(),
        application.getEmail(),
        application.getMessage(),
        application.getStatus(),
        application.getCreatedAt(),
        application.getUpdatedAt());
  }
}
