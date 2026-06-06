package com.normanblog.frestyle.controller;

import com.normanblog.frestyle.dto.CompanyApplicationResponse;
import com.normanblog.frestyle.dto.UpdateApplicationStatusRequest;
import com.normanblog.frestyle.security.CurrentUserProvider;
import com.normanblog.frestyle.service.CompanyApplicationService;
import jakarta.validation.Valid;
import java.util.List;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PatchMapping;
import org.springframework.web.bind.annotation.PathVariable;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/** 企業利用申請の管理エンドポイント(super_admin 専用)。 */
@RestController
@RequestMapping("/api/v2/admin/company-applications")
public class AdminCompanyApplicationController {

  private final CompanyApplicationService applications;
  private final CurrentUserProvider currentUser;

  public AdminCompanyApplicationController(
      CompanyApplicationService applications, CurrentUserProvider currentUser) {
    this.applications = applications;
    this.currentUser = currentUser;
  }

  /** 申請を新しい順で返す。 */
  @GetMapping
  public List<CompanyApplicationResponse> list() {
    currentUser.requireSuperAdmin();

    return applications.listNewestFirst().stream().map(CompanyApplicationResponse::from).toList();
  }

  /** 申請の status を approved / rejected / pending に更新する。 */
  @PatchMapping("/{id}/status")
  public ResponseEntity<Void> updateStatus(
      @PathVariable Long id, @Valid @RequestBody UpdateApplicationStatusRequest request) {
    currentUser.requireSuperAdmin();
    applications.updateStatus(id, request.status());

    return ResponseEntity.noContent().build();
  }
}
