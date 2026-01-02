package com.example.FreStyle.auth;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.oauth2.server.resource.authentication.JwtAuthenticationConverter;
import org.springframework.security.oauth2.server.resource.authentication.JwtGrantedAuthoritiesConverter;
import org.springframework.security.web.SecurityFilterChain;
import org.springframework.security.web.authentication.UsernamePasswordAuthenticationFilter;

import lombok.RequiredArgsConstructor;

import static org.springframework.security.config.Customizer.withDefaults;

@Configuration
@RequiredArgsConstructor
public class SecurityConfig {
    
    // Cognitoの.well-known/jwk.json
    // Spring Security は access_token を自動で decode & validate & set Authentication してくれる
    @Value("${cognito.jwk-set-uri}")
    private final String jwkUri;

    private final JwtCookieFilter jwtCookieFilter;

    
    @Bean
    public SecurityFilterChain securityFilterChain(HttpSecurity http) throws Exception {
        http
        // withDefaultsでは@Configurationで設定したCorsConfigが適用される。
                .cors(withDefaults())
                .csrf(csrf -> csrf.disable())
                .authorizeHttpRequests(auth -> auth
                        .requestMatchers("/api/hello","/api/hello/**","/api/auth/info","/ws/chat/**","/api/auth/**").permitAll()
                        .anyRequest().authenticated())
                .formLogin(form -> form.disable())
                .httpBasic(basic -> basic.disable())
                // httpOnlyCookieのACCESS_TOKENを Authorization: Bearerヘッダーに変換する
                .addFilterBefore(jwtCookieFilter, UsernamePasswordAuthenticationFilter.class)
                .oauth2ResourceServer(oauth2 -> oauth2
                    .jwt(jwt -> jwt
                        .jwtAuthenticationConverter(jwtAuthenticationConverter()) // カスタムコンバーターを作成をする
                        .jwkSetUri(jwkUri)));
        return http.build();
    }
    
    // カスタムでAuthenticationConverterをつくり、audience = client_idでの検証をする
    // これによりIdp、ユーザープールでのログインが実相をできるようになる
    private JwtAuthenticationConverter jwtAuthenticationConverter() {
        // 権限（Scope/Groups）を抽出するコンバーター
        // Cognitoの場合は、通常は'scope'クレームから権限を取得をする
        // またはカスタムクレームからグループを取得
        JwtGrantedAuthoritiesConverter authoritiesConverter = new JwtGrantedAuthoritiesConverter();
        authoritiesConverter.setAuthoritiesClaimName("cognito:groups");
        authoritiesConverter.setAuthorityPrefix("ROLE_");
        
        JwtAuthenticationConverter converter = new JwtAuthenticationConverter();
        converter.setJwtGrantedAuthoritiesConverter(authoritiesConverter);
        converter.setPrincipalClaimName("username");
        
        
        return converter;
        
        // Jwtを認証オブジェクトに変換をし、@AuthenticationPrincipalがnullじゃなくなる
    }

}