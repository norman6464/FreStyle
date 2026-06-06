package com.normanblog.frestyle.service;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import com.normanblog.frestyle.entity.CompanyApplication;
import com.normanblog.frestyle.entity.CompanyApplicationStatus;
import com.normanblog.frestyle.entity.Notification;
import com.normanblog.frestyle.entity.NotificationType;
import com.normanblog.frestyle.entity.Role;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.CompanyApplicationRepository;
import com.normanblog.frestyle.repository.NotificationRepository;
import com.normanblog.frestyle.repository.UserRepository;
import java.util.List;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.web.server.ResponseStatusException;

/** 企業利用申請の作成 / 一覧 / status 更新と、作成時の super_admin 通知を検証する。 */
@SpringBootTest
class CompanyApplicationServiceTest {

  @Autowired private CompanyApplicationService service;
  @Autowired private CompanyApplicationRepository applications;
  @Autowired private UserRepository users;
  @Autowired private NotificationRepository notifications;

  @BeforeEach
  void setUp() {
    notifications.deleteAll();
    applications.deleteAll();
    users.deleteAll();
  }

  @Test
  void create_savesAsPending() {
    CompanyApplication saved = service.create("Acme", "山田", "y@acme.com", "よろしく");

    assertThat(saved.getId()).isNotNull();
    assertThat(saved.getStatus()).isEqualTo(CompanyApplicationStatus.PENDING);
    assertThat(saved.getCompanyName()).isEqualTo("Acme");
  }

  @Test
  void create_notifiesEverySuperAdmin() {
    Long admin1 =
        users
            .save(User.builder().cognitoSub("a1").email("a1@x.com").role(Role.SUPER_ADMIN).build())
            .getId();
    Long admin2 =
        users
            .save(User.builder().cognitoSub("a2").email("a2@x.com").role(Role.SUPER_ADMIN).build())
            .getId();
    // super_admin 以外には通知しない。
    users.save(User.builder().cognitoSub("t1").email("t1@x.com").role(Role.TRAINEE).build());

    service.create("Acme", "山田", "y@acme.com", "よろしく");

    List<Notification> all = notifications.findAll();
    assertThat(all).hasSize(2);
    assertThat(all).allMatch(n -> n.getType().equals(NotificationType.COMPANY_APPLICATION));
    assertThat(all).allMatch(n -> !n.isRead());
    assertThat(all).extracting(Notification::getUserId).containsExactlyInAnyOrder(admin1, admin2);
  }

  @Test
  void create_withoutSuperAdmin_stillSaved() {
    CompanyApplication saved = service.create("Acme", "山田", "y@acme.com", null);

    assertThat(saved.getId()).isNotNull();
    assertThat(notifications.findAll()).isEmpty();
  }

  @Test
  void listNewestFirst_ordersByCreatedAtDesc() {
    service.create("Old", "a", "a@x.com", null);
    service.create("New", "b", "b@x.com", null);

    List<CompanyApplication> rows = service.listNewestFirst();

    assertThat(rows).hasSize(2);
    assertThat(rows.get(0).getCompanyName()).isEqualTo("New");
  }

  @Test
  void updateStatus_validStatus_updates() {
    CompanyApplication app = service.create("Acme", "山田", "y@acme.com", null);

    service.updateStatus(app.getId(), CompanyApplicationStatus.APPROVED);

    assertThat(applications.findById(app.getId()).orElseThrow().getStatus())
        .isEqualTo(CompanyApplicationStatus.APPROVED);
  }

  @Test
  void updateStatus_invalidStatus_throws400() {
    CompanyApplication app = service.create("Acme", "山田", "y@acme.com", null);

    assertThatThrownBy(() -> service.updateStatus(app.getId(), "banana"))
        .isInstanceOf(ResponseStatusException.class)
        .hasMessageContaining("400");
  }

  @Test
  void updateStatus_missingApplication_throws404() {
    assertThatThrownBy(() -> service.updateStatus(999L, CompanyApplicationStatus.APPROVED))
        .isInstanceOf(ResponseStatusException.class)
        .hasMessageContaining("404");
  }
}
