package com.normanblog.frestyle.config;

import org.springframework.boot.context.properties.ConfigurationProperties;

/**
 * コード実行(サンドボックス)の設定。
 *
 * <p>enabled=false なら実行を受け付けず 503 を返す(実行環境の無い場所で安全に無効化できる)。
 */
@ConfigurationProperties(prefix = "frestyle.code-exec")
public record CodeExecProperties(
    Boolean enabled, Long timeoutSeconds, Integer maxOutputBytes, Integer maxCodeBytes) {

  public boolean isEnabled() {
    return enabled == null || enabled; // 既定は有効(runtime に JDK がある前提)。
  }

  public long timeoutSecondsOrDefault() {
    return timeoutSeconds == null || timeoutSeconds <= 0 ? 10L : timeoutSeconds;
  }

  public int maxOutputBytesOrDefault() {
    return maxOutputBytes == null || maxOutputBytes <= 0 ? 64 * 1024 : maxOutputBytes;
  }

  public int maxCodeBytesOrDefault() {
    return maxCodeBytes == null || maxCodeBytes <= 0 ? 64 * 1024 : maxCodeBytes;
  }
}
