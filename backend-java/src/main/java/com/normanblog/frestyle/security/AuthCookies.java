package com.normanblog.frestyle.security;

import jakarta.servlet.http.HttpServletResponse;
import java.time.Duration;
import org.springframework.http.HttpHeaders;
import org.springframework.http.ResponseCookie;
import org.springframework.stereotype.Component;

/**
 * 認証 Cookie(access_token / refresh_token)の発行・破棄を担う。
 *
 * <p>JS から読めないよう HttpOnly、別オリジン(SPA ⇄ API)で送れるよう Secure + SameSite=None。
 * 既存 Go 実装と同じ属性(Path=/、access は expires_in、refresh は 30 日)。
 */
@Component
public class AuthCookies {

  static final String ACCESS_TOKEN = "access_token";
  static final String REFRESH_TOKEN = "refresh_token";
  private static final int ACCESS_DEFAULT_MAX_AGE = 3600;
  private static final Duration REFRESH_MAX_AGE = Duration.ofDays(30);

  public void setAccessToken(HttpServletResponse response, String token, int maxAgeSeconds) {
    int maxAge = maxAgeSeconds > 0 ? maxAgeSeconds : ACCESS_DEFAULT_MAX_AGE;
    add(response, build(ACCESS_TOKEN, token, Duration.ofSeconds(maxAge)));
  }

  public void setRefreshToken(HttpServletResponse response, String token) {
    // Cognito が refresh_token を返さないケース(refresh grant など)では何もしない。
    if (token == null || token.isBlank()) {
      return;
    }
    add(response, build(REFRESH_TOKEN, token, REFRESH_MAX_AGE));
  }

  public void clear(HttpServletResponse response) {
    add(response, build(ACCESS_TOKEN, "", Duration.ZERO));
    add(response, build(REFRESH_TOKEN, "", Duration.ZERO));
  }

  private ResponseCookie build(String name, String value, Duration maxAge) {
    return ResponseCookie.from(name, value)
        .httpOnly(true)
        .secure(true)
        .sameSite("None")
        .path("/")
        .maxAge(maxAge)
        .build();
  }

  private void add(HttpServletResponse response, ResponseCookie cookie) {
    response.addHeader(HttpHeaders.SET_COOKIE, cookie.toString());
  }
}
