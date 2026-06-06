package com.normanblog.frestyle.config;

import org.springframework.boot.context.properties.ConfigurationProperties;

/**
 * Cognito 連携の設定。ECS では COGNITO_* 環境変数から注入される
 * (application.properties で frestyle.cognito.* にマッピング)。
 */
@ConfigurationProperties(prefix = "frestyle.cognito")
public record CognitoProperties(
    String adminGroup, String clientId, String clientSecret, String redirectUri, String tokenUri) {}
