package com.normanblog.frestyle.infra.s3;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import software.amazon.awssdk.core.ResponseBytes;
import software.amazon.awssdk.services.s3.S3Client;
import software.amazon.awssdk.services.s3.model.GetObjectRequest;
import software.amazon.awssdk.services.s3.model.GetObjectResponse;

/** S3 GetObject で添付の実体を取得する本番実装。 */
public class S3AttachmentDownloader implements AttachmentDownloader, AutoCloseable {

  private static final Logger log = LoggerFactory.getLogger(S3AttachmentDownloader.class);

  private final S3Client client;
  private final String bucket;

  public S3AttachmentDownloader(S3Client client, String bucket) {
    this.client = client;
    this.bucket = bucket;
  }

  @Override
  public byte[] download(String key) {
    try {
      ResponseBytes<GetObjectResponse> bytes =
          client.getObjectAsBytes(GetObjectRequest.builder().bucket(bucket).key(key).build());

      return bytes.asByteArray();
    } catch (RuntimeException e) {
      // 取得失敗は致命ではない。添付を落としてテキストだけで送る(部分劣化)。
      log.warn("ai-chat attachment download failed: key={}", key, e);
      return null;
    }
  }

  @Override
  public void close() {
    client.close();
  }
}
