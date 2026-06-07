package com.normanblog.frestyle.service;

import com.normanblog.frestyle.entity.Role;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.CompanyRepository;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Service;
import org.springframework.web.server.ResponseStatusException;

/**
 * trainee が AI チャットを使ってよいかの認可ポリシー。
 *
 * <p>会社の {@code aiChatEnabledForTrainees} が false の場合、その会社の trainee は AI を使えない。
 * company_admin / super_admin は会社設定に関わらず常に利用可。会社未所属(company_id=null)は既定 true。
 */
@Service
public class AiChatAccessPolicy {

  private final CompanyRepository companies;

  public AiChatAccessPolicy(CompanyRepository companies) {
    this.companies = companies;
  }

  /** このユーザーが AI チャットを使えるか(UI のサイドバー表示判定 / me で使う)。 */
  public boolean isEnabledFor(User user) {
    // 管理者は常に利用可(自社の設定を確認するために自分で使う必要がある)。
    if (Role.COMPANY_ADMIN.equals(user.getRole()) || Role.SUPER_ADMIN.equals(user.getRole())) {
      return true;
    }
    if (user.getCompanyId() == null) {
      return true;
    }
    return companies
        .findById(user.getCompanyId())
        .map(c -> c.isAiChatEnabledForTrainees())
        .orElse(true);
  }

  /** AI チャット系エンドポイントの入口で使う。無効なら 403。 */
  public void enforce(User user) {
    if (!isEnabledFor(user)) {
      throw new ResponseStatusException(HttpStatus.FORBIDDEN, "ai_chat_disabled_for_company");
    }
  }
}
