package com.normanblog.frestyle.service;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

import com.normanblog.frestyle.entity.CompanyApplication;
import com.normanblog.frestyle.entity.CompanyApplicationStatus;
import com.normanblog.frestyle.repository.CompanyApplicationRepository;
import java.util.List;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.web.server.ResponseStatusException;

/** 企業利用申請の作成 / 一覧 / status 更新を検証する。 */
@SpringBootTest
class CompanyApplicationServiceTest {

  @Autowired private CompanyApplicationService service;
  @Autowired private CompanyApplicationRepository applications;

  @BeforeEach
  void setUp() {
    applications.deleteAll();
  }

  @Test
  void create_savesAsPending() {
    CompanyApplication saved = service.create("Acme", "山田", "y@acme.com", "よろしく");

    assertThat(saved.getId()).isNotNull();
    assertThat(saved.getStatus()).isEqualTo(CompanyApplicationStatus.PENDING);
    assertThat(saved.getCompanyName()).isEqualTo("Acme");
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
