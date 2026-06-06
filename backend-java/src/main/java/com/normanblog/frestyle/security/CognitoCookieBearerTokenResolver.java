package com.normanblog.frestyle.security;

import jakarta.servlet.http.Cookie;
import jakarta.servlet.http.HttpServletRequest;
import java.util.Set;
import org.springframework.security.oauth2.server.resource.web.BearerTokenResolver;

/**
 * access_token を HttpOnly Cookie から取り出す BearerTokenResolver。
 *
 * <p>標準の Resolver は {@code Authorization: Bearer} ヘッダを見るが、本アプリは XSS で
 * トークンを抜かれないよう HttpOnly Cookie に載せる方式(既存 Go 実装と同じ)を採るため、
 * Cookie から取り出すように差し替える。
 */
public class CognitoCookieBearerTokenResolver implements BearerTokenResolver {

  private static final String COOKIE_NAME = "access_token";

  // 公開エンドポイントでは Bearer 解決をしない(= JWT 検証を走らせない)。
  // とくに /refresh は access_token が期限切れの状態で叩かれるため、ここでトークンを返すと
  // 期限切れ JWT の検証が先に走って 401 になり、refresh 本来の役割(再発行)が壊れる。
  private static final Set<String> PUBLIC_PATHS =
      Set.of(
          "/api/v2/health",
          "/api/v2/auth/login",
          "/api/v2/auth/logout",
          "/api/v2/auth/refresh");

  @Override
  public String resolve(HttpServletRequest request) {
    if (PUBLIC_PATHS.contains(request.getRequestURI())) {
      return null;
    }

    Cookie[] cookies = request.getCookies();
    if (cookies == null) {
      return null;
    }
    for (Cookie cookie : cookies) {
      if (COOKIE_NAME.equals(cookie.getName())) {
        String value = cookie.getValue();
        if (value != null && !value.isBlank()) {
          return value;
        }
      }
    }
    return null;
  }
}
