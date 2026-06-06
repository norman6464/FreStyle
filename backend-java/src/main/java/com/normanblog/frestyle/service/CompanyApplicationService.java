package com.normanblog.frestyle.service;

import com.normanblog.frestyle.entity.CompanyApplication;
import com.normanblog.frestyle.entity.CompanyApplicationStatus;
import com.normanblog.frestyle.entity.NotificationType;
import com.normanblog.frestyle.entity.Role;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.CompanyApplicationRepository;
import com.normanblog.frestyle.repository.UserRepository;
import java.time.Instant;
import java.util.List;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.web.server.ResponseStatusException;

/** 企業利用申請のビジネスロジック(作成 / 一覧 / status 更新)。 */
@Service
public class CompanyApplicationService {

  private static final Logger log = LoggerFactory.getLogger(CompanyApplicationService.class);

  private final CompanyApplicationRepository applications;
  private final UserRepository users;
  private final NotificationService notifications;

  public CompanyApplicationService(
      CompanyApplicationRepository applications,
      UserRepository users,
      NotificationService notifications) {
    this.applications = applications;
    this.users = users;
    this.notifications = notifications;
  }

  /** 公開フォームからの申請を pending で受け付け、全 super_admin へ通知する。 */
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

    CompanyApplication saved = applications.save(application);
    // 申請保存は成功扱いとし、通知作成失敗はログのみ(best-effort)。
    notifySuperAdmins(saved);

    return saved;
  }

  private void notifySuperAdmins(CompanyApplication application) {
    String title = "新しい企業申請が届きました";
    String body =
        String.format(
            "%s（%s / %s）から利用申請がありました。",
            application.getCompanyName(), application.getApplicantName(), application.getEmail());
    for (User admin : users.findByRoleAndDeletedAtIsNull(Role.SUPER_ADMIN)) {
      try {
        notifications.create(admin.getId(), NotificationType.COMPANY_APPLICATION, title, body);
      } catch (RuntimeException e) {
        log.warn("company-application: notify super_admin {} failed", admin.getId(), e);
      }
    }
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
