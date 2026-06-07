package com.normanblog.frestyle.service;

import com.normanblog.frestyle.entity.Company;
import com.normanblog.frestyle.entity.Role;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.CompanyRepository;
import java.time.Instant;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import org.springframework.web.server.ResponseStatusException;

/**
 * 会社単位の設定(trainee への AI 有効化)の取得・更新。
 *
 * <p>操作できるのは company_admin / super_admin のみで、対象は常に自社(自分の company_id)。
 * trainee は設定できない(403)。
 */
@Service
public class CompanySettingsService {

  private final CompanyRepository companies;

  public CompanySettingsService(CompanyRepository companies) {
    this.companies = companies;
  }

  /** 自社の設定を取得する。company_admin / super_admin のみ。 */
  @Transactional(readOnly = true)
  public Company getForAdmin(User actor) {
    requireAdmin(actor);
    return loadOwnCompany(actor);
  }

  /** 自社の AI 有効化フラグを更新する。company_admin / super_admin のみ。 */
  @Transactional
  public Company updateAiChatEnabled(User actor, boolean enabled) {
    requireAdmin(actor);
    Company company = loadOwnCompany(actor);
    company.setAiChatEnabledForTrainees(enabled);
    company.setUpdatedAt(Instant.now());
    return companies.save(company);
  }

  private void requireAdmin(User actor) {
    if (!Role.COMPANY_ADMIN.equals(actor.getRole()) && !Role.SUPER_ADMIN.equals(actor.getRole())) {
      throw new ResponseStatusException(HttpStatus.FORBIDDEN, "forbidden");
    }
  }

  private Company loadOwnCompany(User actor) {
    if (actor.getCompanyId() == null) {
      throw new ResponseStatusException(HttpStatus.BAD_REQUEST, "no_company");
    }
    return companies
        .findById(actor.getCompanyId())
        .orElseThrow(() -> new ResponseStatusException(HttpStatus.NOT_FOUND, "company_not_found"));
  }
}
