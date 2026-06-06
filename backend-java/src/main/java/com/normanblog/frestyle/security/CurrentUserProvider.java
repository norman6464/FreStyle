package com.normanblog.frestyle.security;

import com.normanblog.frestyle.config.CognitoProperties;
import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.repository.UserRepository;
import java.util.List;
import org.springframework.http.HttpStatus;
import org.springframework.security.core.Authentication;
import org.springframework.security.core.context.SecurityContextHolder;
import org.springframework.security.oauth2.jwt.Jwt;
import org.springframework.stereotype.Component;
import org.springframework.web.server.ResponseStatusException;

/**
 * 認証済みリクエストから「アプリ内のユーザー」を解決するヘルパー。
 *
 * <p>検証済み JWT の {@code sub} をキーに users 行を引き、業務処理で使う内部 ID を提供する。
 * JWT は認証済みでも users 行が無い(招待前など)ケースがあるため、その場合は 401 にする。
 */
@Component
public class CurrentUserProvider {

  private final UserRepository users;
  private final String adminGroup;

  public CurrentUserProvider(UserRepository users, CognitoProperties cognito) {
    this.users = users;
    this.adminGroup = cognito.adminGroup();
  }

  /** 認証済み JWT の sub から users 行を引く。未登録なら 401。 */
  public User require() {
    String sub = currentJwt().getSubject();
    return users
        .findByCognitoSubAndDeletedAtIsNull(sub)
        .orElseThrow(() -> new ResponseStatusException(HttpStatus.UNAUTHORIZED, "user_not_found"));
  }

  /** 検証済みの JWT を取り出す。未認証なら 401。 */
  public Jwt currentJwt() {
    Authentication auth = SecurityContextHolder.getContext().getAuthentication();
    if (auth != null && auth.getPrincipal() instanceof Jwt jwt) {
      return jwt;
    }
    throw new ResponseStatusException(HttpStatus.UNAUTHORIZED, "unauthorized");
  }

  /** JWT の cognito:groups を文字列リストで返す(未設定なら空)。 */
  public List<String> groups(Jwt jwt) {
    Object raw = jwt.getClaim("cognito:groups");
    if (raw instanceof List<?> list) {
      return list.stream().filter(String.class::isInstance).map(String.class::cast).toList();
    }
    return List.of();
  }

  /** groups に管理者グループが含まれるか。 */
  public boolean isAdmin(List<String> groups) {
    return groups.contains(adminGroup);
  }
}
