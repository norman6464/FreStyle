package com.normanblog.frestyle.security;

import jakarta.servlet.http.Cookie;
import jakarta.servlet.http.HttpServletRequest;
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

  @Override
  public String resolve(HttpServletRequest request) {
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
