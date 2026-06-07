package com.normanblog.frestyle.infra.s3;

import com.normanblog.frestyle.dto.ProfileImageUploadUrl;

/**
 * AWS を呼ばない stub。bucket 未設定のローカル / テスト環境で使う。
 *
 * <p>本番では {@link S3ProfileImagePresigner} に差し替わる(bucket が設定されている場合)。
 */
public class StubProfileImagePresigner implements ProfileImagePresigner {

  private static final int TTL_SECONDS = 600;
  private static final String DEFAULT_CONTENT_TYPE = "image/png";

  private final String bucket;
  private final String cdnBase;

  public StubProfileImagePresigner(String bucket, String cdnBase) {
    this.bucket = (bucket == null || bucket.isBlank()) ? "stub-bucket" : bucket;
    this.cdnBase = (cdnBase == null) ? "" : cdnBase.replaceAll("/+$", "");
  }

  @Override
  public ProfileImageUploadUrl generate(Long userId, String fileName, String contentType) {
    String type = (contentType == null || contentType.isBlank()) ? DEFAULT_CONTENT_TYPE : contentType;
    String key = ImageKeys.profileKey(userId, fileName, type);
    String uploadUrl = "https://" + bucket + ".s3.amazonaws.com/" + key + "?X-Amz-Stub=1";
    String base = cdnBase.isBlank() ? "https://" + bucket + ".s3.amazonaws.com" : cdnBase;

    return new ProfileImageUploadUrl(uploadUrl, base + "/" + key, key, TTL_SECONDS);
  }
}
