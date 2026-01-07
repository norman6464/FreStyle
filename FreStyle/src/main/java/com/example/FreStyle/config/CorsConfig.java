package com.example.FreStyle.config;

import java.util.Arrays;
import java.util.List;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.cors.CorsConfiguration;
import org.springframework.web.cors.CorsConfigurationSource;
import org.springframework.web.cors.UrlBasedCorsConfigurationSource;
import org.springframework.web.servlet.config.annotation.CorsRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

@Configuration
public class CorsConfig {

  private static final Logger logger = LoggerFactory.getLogger(CorsConfig.class);

  // 許可するオリジンを一元管理
  private static final List<String> ALLOWED_ORIGINS = Arrays.asList(
      "http://fre-style-bucket.s3-website-ap-northeast-1.amazonaws.com",
      "https://dcd3m6lwt0z8u.cloudfront.net",
      "http://localhost:5173",
      "https://normanblog.com",
      "http://normanblog.com"
  );

  private static final List<String> ALLOWED_METHODS = Arrays.asList(
      "GET", "POST", "PUT", "DELETE", "OPTIONS"
  );

  /**
   * Spring Security用のCORS設定（これが重要！）
   * SecurityFilterChainの.cors(withDefaults())がこのBeanを使用する
   */
  @Bean
  public CorsConfigurationSource corsConfigurationSource() {
    logger.info("========== [CorsConfig] CorsConfigurationSource Bean 初期化開始 ==========");
    
    CorsConfiguration configuration = new CorsConfiguration();
    
    // 許可するオリジン
    configuration.setAllowedOrigins(ALLOWED_ORIGINS);
    logger.info("[CorsConfig] 許可オリジン設定: {}", ALLOWED_ORIGINS);
    
    // 許可するHTTPメソッド
    configuration.setAllowedMethods(ALLOWED_METHODS);
    logger.info("[CorsConfig] 許可メソッド設定: {}", ALLOWED_METHODS);
    
    // 許可するヘッダー
    configuration.setAllowedHeaders(List.of("*"));
    logger.info("[CorsConfig] 許可ヘッダー設定: *");
    
    // クレデンシャル（Cookie等）を許可
    configuration.setAllowCredentials(true);
    logger.info("[CorsConfig] allowCredentials: true");
    
    // プリフライトリクエストのキャッシュ時間（秒）
    configuration.setMaxAge(3600L);
    logger.info("[CorsConfig] maxAge: 3600秒");
    
    UrlBasedCorsConfigurationSource source = new UrlBasedCorsConfigurationSource();
    source.registerCorsConfiguration("/**", configuration);
    
    logger.info("========== [CorsConfig] CorsConfigurationSource Bean 初期化完了 ==========");
    return source;
  }

  /**
   * Spring MVC用のCORS設定（後方互換性のため維持）
   */
  @Bean
  public WebMvcConfigurer corsConfigurer() {
    logger.info("[CorsConfig] WebMvcConfigurer (MVC用CORS) Bean 初期化");
    
    return new WebMvcConfigurer() {

      @Override
      public void addCorsMappings(CorsRegistry registry) {
        logger.info("[CorsConfig] addCorsMappings() 実行 - MVC用CORS設定適用");
        registry.addMapping("/**")
            .allowedOrigins(ALLOWED_ORIGINS.toArray(new String[0]))
            .allowCredentials(true)
            .allowedMethods(ALLOWED_METHODS.toArray(new String[0]))
            .allowedHeaders("*");
      }

    };
  }

}