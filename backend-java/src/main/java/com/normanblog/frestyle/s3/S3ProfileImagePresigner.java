package com.normanblog.frestyle.s3;

import com.normanblog.frestyle.dto.ProfileImageUploadUrl;
import java.time.Duration;
import software.amazon.awssdk.services.s3.model.PutObjectRequest;
import software.amazon.awssdk.services.s3.presigner.S3Presigner;
import software.amazon.awssdk.services.s3.presigner.model.PresignedPutObjectRequest;
import software.amazon.awssdk.services.s3.presigner.model.PutObjectPresignRequest;

/**
 * AWS S3 への PutObject 署名付き URL を発行する本番実装。
 *
 * <p>contentType は署名に焼き込まれるため、クライアントの PUT ヘッダと完全一致が必要
 * (不一致だと SignatureDoesNotMatch)。配信用 URL は CDN ベース + key で組み立てる。
 */
public class S3ProfileImagePresigner implements ProfileImagePresigner {

  private static final Duration TTL = Duration.ofMinutes(10);
  private static final String DEFAULT_CONTENT_TYPE = "image/png";

  private final S3Presigner presigner;
  private final String bucket;
  private final String cdnBase;

  public S3ProfileImagePresigner(S3Presigner presigner, String bucket, String cdnBase) {
    this.presigner = presigner;
    this.bucket = bucket;
    this.cdnBase = trimTrailingSlash(cdnBase);
  }

  @Override
  public ProfileImageUploadUrl generate(Long userId, String fileName, String contentType) {
    String type = (contentType == null || contentType.isBlank()) ? DEFAULT_CONTENT_TYPE : contentType;
    String key = ImageKeys.profileKey(userId, fileName, type);

    PutObjectRequest objectRequest =
        PutObjectRequest.builder().bucket(bucket).key(key).contentType(type).build();
    PutObjectPresignRequest presignRequest =
        PutObjectPresignRequest.builder()
            .signatureDuration(TTL)
            .putObjectRequest(objectRequest)
            .build();
    PresignedPutObjectRequest presigned = presigner.presignPutObject(presignRequest);

    return new ProfileImageUploadUrl(
        presigned.url().toString(), cdnBase + "/" + key, key, (int) TTL.getSeconds());
  }

  private static String trimTrailingSlash(String s) {
    if (s == null) {
      return "";
    }
    int end = s.length();
    while (end > 0 && s.charAt(end - 1) == '/') {
      end--;
    }
    return s.substring(0, end);
  }
}
