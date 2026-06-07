package com.normanblog.frestyle.config;

import org.springframework.boot.context.properties.ConfigurationProperties;

/**
 * AI チャットの Bedrock 設定。ECS では環境変数から注入される
 * (application.properties で frestyle.bedrock.* にマッピング)。
 *
 * <p>modelId が空のときは Bedrock を呼ばない stub client にフォールバックする(ローカル / テスト)。
 */
@ConfigurationProperties(prefix = "frestyle.bedrock")
public record BedrockProperties(String modelId, String region) {

  public String regionOrDefault() {
    return region == null || region.isBlank() ? "ap-northeast-1" : region.trim();
  }
}
