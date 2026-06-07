package com.normanblog.frestyle.s3;

import com.normanblog.frestyle.dto.AiChatAttachmentUploadUrl;
import java.time.Duration;
import software.amazon.awssdk.services.s3.model.PutObjectRequest;
import software.amazon.awssdk.services.s3.presigner.S3Presigner;
import software.amazon.awssdk.services.s3.presigner.model.PresignedPutObjectRequest;
import software.amazon.awssdk.services.s3.presigner.model.PutObjectPresignRequest;

/**
 * AI チャット添付の S3 PutObject 署名付き URL を発行する本番実装。
 *
 * <p>contentType は署名に焼き込まれるため、クライアントの PUT ヘッダと完全一致が必要。
 */
public class S3AiChatAttachmentPresigner implements AiChatAttachmentPresigner, AutoCloseable {

  private static final Duration TTL = Duration.ofMinutes(10);

  private final S3Presigner presigner;
  private final String bucket;

  public S3AiChatAttachmentPresigner(S3Presigner presigner, String bucket) {
    this.presigner = presigner;
    this.bucket = bucket;
  }

  @Override
  public AiChatAttachmentUploadUrl generate(Long userId, String filename, String contentType) {
    String key = ImageKeys.aiChatAttachmentKey(userId, filename, contentType);

    PutObjectRequest objectRequest =
        PutObjectRequest.builder().bucket(bucket).key(key).contentType(contentType).build();
    PutObjectPresignRequest presignRequest =
        PutObjectPresignRequest.builder()
            .signatureDuration(TTL)
            .putObjectRequest(objectRequest)
            .build();
    PresignedPutObjectRequest presigned = presigner.presignPutObject(presignRequest);

    return new AiChatAttachmentUploadUrl(
        presigned.url().toString(), key, (int) TTL.getSeconds());
  }

  // S3Presigner は Closeable。Spring の destroy-method 推論でコンテキスト終了時に閉じる。
  @Override
  public void close() {
    presigner.close();
  }
}
