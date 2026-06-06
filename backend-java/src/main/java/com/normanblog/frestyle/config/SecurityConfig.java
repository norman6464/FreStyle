package com.normanblog.frestyle.config;

import com.normanblog.frestyle.security.CognitoCookieBearerTokenResolver;
import org.springframework.context.annotation.Bean;
import org.springframework.security.config.Customizer;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.config.annotation.web.configurers.AbstractHttpConfigurer;
import org.springframework.security.config.http.SessionCreationPolicy;
import org.springframework.security.oauth2.server.resource.web.BearerTokenResolver;
import org.springframework.security.web.SecurityFilterChain;

/** Cognito の access_token(JWT)を JWKS 検証する Spring Security 設定。 */
@Configuration
public class SecurityConfig {

  // ヘルスチェックは LB が認証なしで叩くため誰でも通す。それ以外は JWT 必須。
  @Bean
  SecurityFilterChain securityFilterChain(HttpSecurity http, BearerTokenResolver bearerTokenResolver)
      throws Exception {
    http
        // JWT を Cookie で運ぶステートレス API のため、セッションは持たない。
        .sessionManagement(sm -> sm.sessionCreationPolicy(SessionCreationPolicy.STATELESS))
        // TODO(認証フォローアップ): Cookie 認証向けの CSRF 対策(Go の CsrfMiddleware 相当)を追加する。
        .csrf(AbstractHttpConfigurer::disable)
        .authorizeHttpRequests(
            auth ->
                auth.requestMatchers("/api/v2/health")
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
}
