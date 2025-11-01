package com.example.FreStyle.auth;

import jakarta.servlet.FilterChain;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.Cookie;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.stereotype.Component;
import org.springframework.web.filter.OncePerRequestFilter;

import java.io.IOException;

// 1リクエスト1回のみ実行するクラスを継承をしている
@Component
public class JwtCookieFilter extends OncePerRequestFilter {

  @Override
  protected void doFilterInternal(HttpServletRequest request,
      HttpServletResponse response,
      FilterChain filterChain)
      throws ServletException, IOException {

    HttpServletRequest modifiedRequest = request;

    // すでにAuthorizationヘッダーがある場合は何もしない（念のため）
    if (request.getHeader("Authorization") == null && request.getCookies() != null) {
      for (Cookie cookie : request.getCookies()) {
        if ("ACCESS_TOKEN".equals(cookie.getName())) {
          String bearerToken = "Bearer " + cookie.getValue();
          modifiedRequest = new HttpServletRequestWrapperWithHeader(request, "Authorization", bearerToken);
          break;
        }
      }
    }

    // 次のフィルターにラップ済みのリクエストをわたす
    filterChain.doFilter(modifiedRequest, response);
  }
}
