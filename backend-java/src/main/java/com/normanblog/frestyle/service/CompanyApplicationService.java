package com.normanblog.frestyle.service;

import com.normanblog.frestyle.entity.CompanyApplication;
import com.normanblog.frestyle.entity.CompanyApplicationStatus;
import com.normanblog.frestyle.repository.CompanyApplicationRepository;
import java.time.Instant;
import java.util.List;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.web.server.ResponseStatusException;

/** 企業利用申請のビジネスロジック(作成 / 一覧 / status 更新)。 */
@Service
public class CompanyApplicationService {

  private final CompanyApplicationRepository applications;

  public CompanyApplicationService(CompanyApplicationRepository applications) {
    this.applications = applications;
  }

  /** 公開フォームからの申請を pending で受け付ける。 */
  public CompanyApplication create(
      String companyName, String applicantName, String email, String message) {
    Instant now = Instant.now();
    CompanyApplication application =
        CompanyApplication.builder()
            .companyName(companyName)
            .applicantName(applicantName)
            .email(email)
            .message(message)
            .status(CompanyApplicationStatus.PENDING)
            .createdAt(now)
            .updatedAt(now)
            .build();

    return applications.save(application);
  }

  /** 申請を新しい順に返す(super_admin 用)。 */
  public List<CompanyApplication> listNewestFirst() {
    return applications.findAllByOrderByCreatedAtDesc();
  }

  /** 申請の status を更新する。status が許容値でなければ 400、申請が無ければ 404。 */
  public void updateStatus(Long id, String status) {
    if (!CompanyApplicationStatus.ALL.contains(status)) {
      throw new ResponseStatusException(HttpStatus.BAD_REQUEST, "invalid status");
    }

    CompanyApplication application =
        applications
            .findById(id)
            .orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND, "not found"));

    application.setStatus(status);
    application.setUpdatedAt(Instant.now());
    applications.save(application);
  }
}
