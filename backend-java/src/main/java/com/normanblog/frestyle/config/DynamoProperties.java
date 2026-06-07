package com.normanblog.frestyle.config;

import org.springframework.boot.context.properties.ConfigurationProperties;

/**
 * AI チャットメッセージの DynamoDB 設定。ECS では環境変数から注入される
 * (application.properties で frestyle.dynamo.* にマッピング)。
 *
 * <p>aiChatTable が空のときは DynamoDB を呼ばない stub reader にフォールバックする(ローカル / テスト)。
 */
@ConfigurationProperties(prefix = "frestyle.dynamo")
public record DynamoProperties(String aiChatTable, String region) {

  public String regionOrDefault() {
    return region == null || region.isBlank() ? "ap-northeast-1" : region.trim();
  }
}
