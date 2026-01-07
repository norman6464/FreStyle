package com.example.FreStyle.config;

import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.CorsRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

@Configuration
public class CorsConfig {

  @Bean
  public WebMvcConfigurer corsConfigurer() {
    return new WebMvcConfigurer() {

      @Override
      public void addCorsMappings(CorsRegistry registry) {
        registry.addMapping("/**") // すべてのエンドポイントに適用
            .allowedOrigins(
                "http://fre-style-bucket.s3-website-ap-northeast-1.amazonaws.com",
                "https://dcd3m6lwt0z8u.cloudfront.net",
                "http://localhost:5173",
                "https://normanblog.com"
            )
            .allowCredentials(true) // Cookieや認証情報を扱う場合はtrue
            .allowedMethods("GET", "POST", "PUT", "DELETE", "OPTIONS")
            .allowedHeaders("*");
      }

    };
  }

}