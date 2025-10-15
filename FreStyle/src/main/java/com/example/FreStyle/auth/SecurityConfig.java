package com.example.FreStyle.auth;

import javax.swing.Spring;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.web.SecurityFilterChain;

@Configuration
public class SecurityConfig {
    
    // Cognitoの.well-known/jwk.json
    // Spring Security は access_token を自動で decode & validate & set Authentication してくれる
    @Value("${cognito.jwk-set-uri}")
    private String jwkUri;

    
    @Bean
    public SecurityFilterChain securityFilterChain(HttpSecurity http) throws Exception {
        http
                .csrf(csrf -> csrf.disable()) // CSRF無効（REST APIの実装のため）
                .authorizeHttpRequests(auth -> auth
                        .requestMatchers("/api/auth/**").permitAll() // 認証関連は公開
                        .anyRequest().authenticated()) 

                .formLogin(form -> form.disable()) // フォームログイン無効
                .httpBasic(basic -> basic.disable()) // Basic認証無効
                
                // JWTのデコード、署名と期限の検証および必要な権限があるかを確保するスコープクレームの確認が含まれる。
                .oauth2ResourceServer(oauth2 -> oauth2
                    .jwt(jwt -> jwt
                        .jwkSetUri(jwkUri)));

        return http.build();
    }

}