package com.normanblog.frestyle.config;

import com.normanblog.frestyle.security.AiChatAccessInterceptor;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.InterceptorRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

/** Spring MVC の横断設定。AI チャット系に会社設定ベースのアクセス制御を噛ませる。 */
@Configuration
public class WebConfig implements WebMvcConfigurer {

  private final AiChatAccessInterceptor aiChatAccessInterceptor;

  public WebConfig(AiChatAccessInterceptor aiChatAccessInterceptor) {
    this.aiChatAccessInterceptor = aiChatAccessInterceptor;
  }

  @Override
  public void addInterceptors(InterceptorRegistry registry) {
    registry.addInterceptor(aiChatAccessInterceptor).addPathPatterns("/api/v2/ai-chat/**");
  }
}
