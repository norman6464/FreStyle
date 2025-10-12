package com.example.FreStyle.auth;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.web.SecurityFilterChain;

@Configuration
public class SecurityConfig {

    @Bean
    public SecurityFilterChain securityFilterChain(HttpSecurity http) throws Exception {
        http
                .csrf(csrf -> csrf.disable()) // CSRF無効（REST APIの実装のため）
                .authorizeHttpRequests(auth -> auth
                        .anyRequest().permitAll()) // 全リクエスト許可

                .formLogin(form -> form.disable()) // フォームログイン無効
                .httpBasic(basic -> basic.disable()) // Basic認証無効
                .oauth2ResourceServer(oauth2 -> oauth2.disable()); // Oauth2リソースサーバー無効

        return http.build();
    }

}