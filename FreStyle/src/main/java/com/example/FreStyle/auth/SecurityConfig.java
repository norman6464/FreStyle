package com.example.FreStyle.auth;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.security.config.annotation.web.builders.HttpSecurity;
import org.springframework.security.web.SecurityFilterChain;
import static org.springframework.security.config.Customizer.withDefaults;

@Configuration
public class SecurityConfig {
    
    // Cognitoの.well-known/jwk.json
    // Spring Security は access_token を自動で decode & validate & set Authentication してくれる
    @Value("${cognito.jwk-set-uri}")
    private String jwkUri;

    
    @Bean
    public SecurityFilterChain securityFilterChain(HttpSecurity http) throws Exception {
        http
        // withDefaultsでは@Configurationで設定したCorsConfigが適用される。
                .cors(withDefaults())
                .csrf(csrf -> csrf.disable())
                .authorizeHttpRequests(auth -> auth
                        .requestMatchers("/api/auth/**").permitAll()
                        .anyRequest().authenticated()) 
                .formLogin(form -> form.disable())
                .httpBasic(basic -> basic.disable()) 
                .oauth2ResourceServer(oauth2 -> oauth2
                    .jwt(jwt -> jwt
                        .jwkSetUri(jwkUri)));

        return http.build();
    }

}