package com.normanblog.frestyle.security;

import com.normanblog.frestyle.entity.User;
import com.normanblog.frestyle.service.AiChatAccessPolicy;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.http.HttpMethod;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.HandlerInterceptor;

/**
 * AI チャット系エンドポイントの入口で、会社設定により trainee の利用を遮断する。
 *
 * <p>UI 非表示だけに頼らず、API 直叩きもサーバ側で 403 にするための認可境界。
 */
@Component
public class AiChatAccessInterceptor implements HandlerInterceptor {

  private final CurrentUserProvider currentUser;
  private final AiChatAccessPolicy policy;

  public AiChatAccessInterceptor(CurrentUserProvider currentUser, AiChatAccessPolicy policy) {
    this.currentUser = currentUser;
    this.policy = policy;
  }

  @Override
  public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) {
    // CORS preflight(OPTIONS)は認可対象外。
    if (HttpMethod.OPTIONS.matches(request.getMethod())) {
      return true;
    }
    User user = currentUser.require(); // 未認証は 401
    policy.enforce(user); // 無効会社の trainee は 403
    return true;
  }
}
