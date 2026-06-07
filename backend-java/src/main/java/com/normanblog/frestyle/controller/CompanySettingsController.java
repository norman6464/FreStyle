package com.normanblog.frestyle.controller;

import com.normanblog.frestyle.dto.CompanySettingsResponse;
import com.normanblog.frestyle.dto.UpdateCompanySettingsRequest;
import com.normanblog.frestyle.entity.Company;
import com.normanblog.frestyle.security.CurrentUserProvider;
import com.normanblog.frestyle.service.CompanySettingsService;
import jakarta.validation.Valid;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.PutMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

/**
 * 会社設定 API。company_admin / super_admin が自社の設定(trainee への AI 有効化)を取得・更新する。
 */
@RestController
@RequestMapping("/api/v2/company/settings")
public class CompanySettingsController {

  private final CompanySettingsService settings;
  private final CurrentUserProvider currentUser;

  public CompanySettingsController(
      CompanySettingsService settings, CurrentUserProvider currentUser) {
    this.settings = settings;
    this.currentUser = currentUser;
  }

  @GetMapping
  public CompanySettingsResponse get() {
    Company company = settings.getForAdmin(currentUser.require());
    return new CompanySettingsResponse(company.isAiChatEnabledForTrainees());
  }

  @PutMapping
  public CompanySettingsResponse update(@Valid @RequestBody UpdateCompanySettingsRequest request) {
    Company company =
        settings.updateAiChatEnabled(currentUser.require(), request.aiChatEnabledForTrainees());
    return new CompanySettingsResponse(company.isAiChatEnabledForTrainees());
  }
}
