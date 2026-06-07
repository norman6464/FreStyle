package com.normanblog.frestyle.config;

import org.springframework.boot.context.properties.ConfigurationProperties;

/**
 * 画像アップロード(S3 PUT 署名付き URL)の設定。ECS では環境変数から注入される
 * (application.properties で frestyle.s3.* にマッピング)。
 *
 * <p>bucket が空のときは AWS を呼ばない stub presigner にフォールバックする(ローカル / テスト用)。
 */
@ConfigurationProperties(prefix = "frestyle.s3")
public record S3Properties(String bucket, String cdnBase, String region) {

  /** region 未設定でも presigner が成立するよう既定リージョンにフォールバックする。 */
  public String regionOrDefault() {
    return region == null || region.isBlank() ? "ap-northeast-1" : region.trim();
  }
}
