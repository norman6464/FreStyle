package com.example.FreStyle.config;

import java.util.Arrays;
import java.util.List;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.cors.CorsConfiguration;
import org.springframework.web.cors.CorsConfigurationSource;
import org.springframework.web.cors.UrlBasedCorsConfigurationSource;
import org.springframework.web.servlet.config.annotation.CorsRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

@Configuration
public class CorsConfig {

  /**
   * Spring Security用のCORS設定
   * SecurityFilterChainより前にCORSを処理するために必要
   */
  @Bean
  public CorsConfigurationSource corsConfigurationSource() {
    CorsConfiguration configuration = new CorsConfiguration();
    
    // 許可するオリジン
    configuration.setAllowedOrigins(Arrays.asList(
        "http://localhost:5173",
        "https://normanblog.com"
    ));
    
    // 許可するHTTPメソッド
    configuration.setAllowedMethods(Arrays.asList(
        "GET", "POST", "PUT", "DELETE", "OPTIONS"
    ));
    
    // 許可するヘッダー
    configuration.setAllowedHeaders(List.of("*"));
    
    // クレデンシャル（Cookie等）を許可
    configuration.setAllowCredentials(true);
    
    // プリフライトリクエストのキャッシュ時間（秒）
    configuration.setMaxAge(3600L);
    
    UrlBasedCorsConfigurationSource source = new UrlBasedCorsConfigurationSource();
    source.registerCorsConfiguration("/**", configuration);
    
    return source;
  }

  /**
   * Spring MVC用のCORS設定（後方互換性のため維持）
   */
  @Bean
  public WebMvcConfigurer corsConfigurer() {
    return new WebMvcConfigurer() {

      @Override
      public void addCorsMappings(CorsRegistry registry) {
        registry.addMapping("/**")
            .allowedOrigins(
                "http://localhost:5173",
                "https://normanblog.com"
            )
            .allowCredentials(true)
            .allowedMethods("GET", "POST", "PUT", "DELETE", "OPTIONS")
            .allowedHeaders("*");
      }

    };
  }

}
