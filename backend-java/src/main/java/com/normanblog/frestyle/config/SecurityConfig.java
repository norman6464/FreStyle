package com.normanblog.frestyle.config;

import com.normanblog.frestyle.security.CognitoCookieBearerTokenResolver;
import java.util.Arrays;
import java.util.List;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.http.HttpMethod;
import org.springframework.security.config.Customizer;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configurers.AbstractHttpConfigurer;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.oauth2.server.resource.web.BearerTokenResolver;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.web.cors.CorsConfiguration;
import org.springframework.web.cors.CorsConfigurationSource;
import org.springframework.web.cors.UrlBasedCorsConfigurationSource;

/** Cognito の access_token(JWT)を JWKS 検証する Spring Security 設定。 */
@Configuration
public class SecurityConfig {

  // CORS 許可オリジン(Go 版 CorsMiddleware と同じ allowlist)。env FRESTYLE_CORS_ALLOWED_ORIGINS で上書き可。
  // 末尾スラッシュ無しの完全一致。credentials を返すため "*" は使わない。
  @Value(
      "${frestyle.cors.allowed-origins:"
          + "https://normanblog.com,http://normanblog.com,http://localhost:5173,"
          + "https://dcd3m6lwt0z8u.cloudfront.net,"
          + "http://fre-style-bucket.s3-website-ap-northeast-1.amazonaws.com}")
  private String allowedOrigins;

  // ヘルスチェックは LB が認証なしで叩くため誰でも通す。それ以外は JWT 必須。
  @Bean
  SecurityFilterChain securityFilterChain(HttpSecurity http, BearerTokenResolver bearerTokenResolver)
      throws Exception {
    http
        // フロント(normanblog.com)と API(api.normanblog.com)は別オリジンのため CORS を有効化する。
        .cors(Customizer.withDefaults())
        // JWT を Cookie で運ぶステートレス API のため、セッションは持たない。
        .sessionManagement(sm -> sm.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
        // TODO(認証フォローアップ): Cookie 認証向けの CSRF 対策(Go の CsrfMiddleware 相当)を追加する。
        .csrf(AbstractHttpConfigurer::disable)
        .authorizeHttpRequests(
            auth ->
                auth.requestMatchers("/api/v2/health")
                    .permitAll()
                    // ログインフロー(login/logout/refresh)は Cookie を発行・破棄する側なので
                    // JWT 検証の前段。認証不要で通す。
                    .requestMatchers("/api/v2/auth/login", "/api/v2/auth/logout", "/api/v2/auth/refresh")
                    .permitAll()
                    // 企業利用申請の作成はログイン前の未登録ユーザーが使う公開フォーム。
                    .requestMatchers(HttpMethod.POST, "/api/v2/company-applications")
                    .permitAll()
                    // 例外発生時に Spring Boot が /error へフォワードする。ここが認証必須だと
                    // バリデーション 400 等が 401 に化けるため公開する(本来の status を返させる)。
                    .requestMatchers("/error")
                    .permitAll()
                    .anyRequest()
                    .authenticated())
        .oauth2ResourceServer(
            oauth2 ->
                oauth2.bearerTokenResolver(bearerTokenResolver).jwt(Customizer.withDefaults()));
    return http.build();
  }

  @Bean
  BearerTokenResolver cognitoCookieBearerTokenResolver() {
    return new CognitoCookieBearerTokenResolver();
  }

  /**
   * CORS 設定。フロント(別オリジン)から Cookie 付き(credentials)で API を叩くため、許可オリジンは
   * allowlist の完全一致にし、ワイルドカードは使わない(credentials と "*" は併用不可)。
   */
  @Bean
  CorsConfigurationSource corsConfigurationSource() {
    CorsConfiguration config = new CorsConfiguration();
    config.setAllowedOrigins(
        Arrays.stream(allowedOrigins.split(",")).map(String::trim).filter(s -> !s.isEmpty()).toList());
    config.setAllowedMethods(List.of("GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"));
    config.setAllowedHeaders(List.of("Content-Type", "Authorization", "X-Requested-With"));
    config.setAllowCredentials(true);
    config.setMaxAge(3600L);

    UrlBasedCorsConfigurationSource source = new UrlBasedCorsConfigurationSource();
    source.registerCorsConfiguration("/**", config);
    return source;
  }
}
